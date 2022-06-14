package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/emerishq/demeris-api-server/api/account"
	"github.com/getsentry/sentry-go"
)

func (a *App) Balances(ctx context.Context, addresses []string) ([]account.Balance, error) {
	defer sentry.StartSpan(ctx, "usecase.Balances").Finish()

	if len(addresses) == 0 {
		return []account.Balance{}, nil
	}
	balances, err := a.db.Balances(ctx, addresses)
	if err != nil {
		return nil, err
	}
	if len(balances) == 0 {
		return nil, fmt.Errorf("balances not found for addresses %v", addresses)
	}
	verifiedDenoms, err := a.verifiedDenomsMap(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: get unique chains
	// perhaps we can remove this since there will be another endpoint specifically for fee tokens
	res := make([]account.Balance, 0, len(balances))
	for _, b := range balances {
		balance := account.Balance{
			Address:   b.Address,
			Amount:    b.Amount,
			OnChain:   b.ChainName,
			Verified:  verifiedDenoms[b.Denom],
			BaseDenom: b.Denom,
		}

		if strings.HasPrefix(b.Denom, "ibc/") {
			// is ibc token
			balance.Ibc = account.IbcInfo{
				Hash: b.Denom[4:],
			}
			// if err is nil, the ibc denom has a denom trace associated with it
			// so we return it, along with its verified status as well as the complete ibc
			// path
			// otherwise, since we don't touch `verified` and `baseDenom` variables, we stick to the
			// original `ibc/...` denom, which will be unverified by default
			denomTrace, err := a.db.DenomTrace(ctx, b.ChainName, b.Denom[4:])
			if err == nil {
				balance.Ibc.Path = denomTrace.Path
				balance.BaseDenom = denomTrace.BaseDenom
				balance.Verified = verifiedDenoms[denomTrace.BaseDenom]
			}
		}

		res = append(res, balance)
	}
	return res, nil
}

func (a *App) verifiedDenomsMap(ctx context.Context) (map[string]bool, error) {
	chains, err := a.db.VerifiedDenoms(ctx)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]bool)
	for _, cc := range chains {
		for _, vd := range cc {
			ret[vd.Name] = vd.Verified
		}
	}
	return ret, nil
}
