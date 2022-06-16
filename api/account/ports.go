package account

import (
	"context"
)

//go:generate mockgen -package account_test -source ports.go -destination ports_mocks_test.go

type App interface {
	DeriveRawAddress(ctx context.Context, rawAddress string) ([]string, error)
	Balances(ctx context.Context, addresses []string) ([]Balance, error)
	UnbondingDelegations(ctx context.Context, addresses []string) ([]UnbondingDelegation, error)
	StakingBalances(ctx context.Context, addresses []string) ([]StakingBalance, error)
}
