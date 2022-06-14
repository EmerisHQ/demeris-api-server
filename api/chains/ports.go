package chains

import (
	"context"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/emerishq/demeris-backend-models/cns"
)

//go:generate mockgen -package chains_test -source ports.go -destination ports_mocks_test.go

type App interface {
	StakingAPR(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error)
}

type CacheBackend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
}
