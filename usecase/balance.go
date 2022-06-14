package usecase

import (
	"context"
	"fmt"

	"github.com/emerishq/demeris-api-server/api/account"
	"github.com/emerishq/demeris-backend-models/tracelistener"
)

func (a *App) Balances(ctx context.Context, addresses []string) ([]account.Balance, error) {
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
	vd, err := a.verifiedDenomsMap(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: get unique chains
	// perhaps we can remove this since there will be another endpoint specifically for fee tokens
	res := make([]account.Balance, len(balances))
	for i, b := range balances {
		res[i] = a.balanceRespForBalance(
			ctx,
			b,
			vd,
		)
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

	return ret, err
}

func (a *App) balanceRespForBalance(ctx context.Context, rawBalance tracelistener.BalanceRow, vd map[string]bool) account.Balance {
	balance := account.Balance{
		Address: rawBalance.Address,
		Amount:  rawBalance.Amount,
		OnChain: rawBalance.ChainName,
	}

	verified := vd[rawBalance.Denom]
	baseDenom := rawBalance.Denom

	if rawBalance.Denom[:4] == "ibc/" {
		// is ibc token
		balance.Ibc = account.IbcInfo{
			Hash: rawBalance.Denom[4:],
		}

		// if err is nil, the ibc denom has a denom trace associated with it
		// so we return it, along with its verified status as well as the complete ibc
		// path

		// otherwise, since we don't touch `verified` and `baseDenom` variables, we stick to the
		// original `ibc/...` denom, which will be unverified by default
		denomTrace, err := a.db.DenomTrace(ctx, rawBalance.ChainName, rawBalance.Denom[4:])
		if err == nil {
			balance.Ibc.Path = denomTrace.Path
			baseDenom = denomTrace.BaseDenom
			verified = vd[denomTrace.BaseDenom]
		}
	}

	balance.Verified = verified
	balance.BaseDenom = baseDenom

	return balance
}
