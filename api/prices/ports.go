package prices

import (
	"context"

	"github.com/emerishq/demeris-api-server/usecase"
)

type App interface {
	PoolPrices(ctx context.Context) usecase.PoolPricesResult
}
