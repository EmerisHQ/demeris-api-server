package chains

import (
	"context"
	"time"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
)

//go:generate mockgen -package chains_test -source ports.go -destination ports_mocks_test.go

type App interface {
	StakingAPR(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error)
}

type CacheBackend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
}
