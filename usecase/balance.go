package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/emerishq/demeris-api-server/api/account"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
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
		return []account.Balance{}, nil
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

// StakingBalance returns the staking balances of addresses.
func (a *App) StakingBalances(ctx context.Context, addresses []string) ([]account.StakingBalance, error) {
	defer sentry.StartSpan(ctx, "usecase.StakingBalances").Finish()

	if len(addresses) == 0 {
		return []account.StakingBalance{}, nil
	}

	delegations, err := a.db.Delegations(ctx, addresses)
	if err != nil {
		return nil, err
	}
	if len(delegations) == 0 {
		return []account.StakingBalance{}, nil
	}

	var res []account.StakingBalance
	for _, del := range delegations {
		delegationAmount, err := sdktypes.NewDecFromStr(del.Amount)
		if err != nil {
			return nil, fmt.Errorf("cannot convert delegation amount to Dec: %w", err)
		}

		validatorShares, err := sdktypes.NewDecFromStr(del.ValidatorShares)
		if err != nil {
			return nil, fmt.Errorf("cannot convert validator total shares to Dec: %w", err)
		}

		validatorTokens, err := sdktypes.NewDecFromStr(del.ValidatorTokens)
		if err != nil {
			return nil, fmt.Errorf("cannot convert validator total tokens to Dec: %w", err)
		}

		// apply shares * total_validator_balance / total_validator_shares
		balance := delegationAmount.Mul(validatorTokens).Quo(validatorShares)
		res = append(res, account.StakingBalance{
			ValidatorAddress: del.Validator,
			Amount:           balance.String(),
			ChainName:        del.ChainName,
		})
	}
	return res, nil
}

func (a *App) UnbondingDelegations(ctx context.Context, addresses []string) ([]account.UnbondingDelegation, error) {
	defer sentry.StartSpan(ctx, "usecase.UnbondingDelegations").Finish()

	if len(addresses) == 0 {
		return []account.UnbondingDelegation{}, nil
	}

	unbondings, err := a.db.UnbondingDelegations(ctx, addresses)
	if err != nil {
		return nil, err
	}
	res := make([]account.UnbondingDelegation, len(unbondings))
	for i, unbonding := range unbondings {
		res[i] = account.UnbondingDelegation{
			ValidatorAddress: unbonding.Validator,
			Entries:          unbonding.Entries,
			ChainName:        unbonding.ChainName,
		}
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
