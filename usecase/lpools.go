package usecase

import (
	"context"
	"errors"
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/emerishq/demeris-api-server/lib/options"
)

type StrategyID string

// PoolsStrategy is the interface that defines how to retrieve a list of
// supported liquidity pools.
//
// Each strategy must have a unique ID so the pools can be easily categorized.
type PoolsStrategy interface {
	ID() StrategyID
	Pools(ctx context.Context) ([]LiquidityPool, error)
}

// PoolPricesResult is a map containing the results of multiple PoolsStrategies.
//
// Example:
//     "osmosis"    -> [ {pool1}, {error pool2}, {pool3} ]
//     "crescent"   -> [ {pool1}, {pool2}, {pool3} ]
//     "chainx"     -> {error chainx}
type PoolPricesResult map[StrategyID]options.O[[]options.O[PoolPrice]]

func (a *App) PoolPrices(ctx context.Context) PoolPricesResult {
	crescent := NewCrescentPoolsStrategy(a.crescentClient, a.sdkServiceClients)
	osmosis := NewOsmosisPoolsStrategy(a.osmosisClient)

	strategies := []PoolsStrategy{
		crescent,
		osmosis,
	}

	result := make(PoolPricesResult, len(strategies))
	for _, strategy := range strategies {
		result[strategy.ID()] = options.Wrap(a.execPoolsStrategy(ctx, strategy))
	}

	return result
}

func (a *App) execPoolsStrategy(ctx context.Context, s PoolsStrategy) ([]options.O[PoolPrice], error) {
	// list pools
	pools, err := s.Pools(ctx)
	if err != nil {
		err = fmt.Errorf("error in pools strategy %s: %w", s.ID(), err)
		return nil, err
	}

	// get prices for each pool
	res := make([]options.O[PoolPrice], 0, len(pools))
	for _, pool := range pools {
		poolPrice, err := NewPoolPrice(ctx, pool, a.denomPricer)
		res = append(res, options.Wrap(poolPrice, err))
	}

	// ignore pools with "expected" errors
	res = filterByError(res, ErrDenomNotFound)
	res = filterByError(res, ErrIBCTraceNotFound)

	return res, nil
}

// filterByError takes a slice of options and returns a new slice of options
// filtering out all the options that matched the given error kind.
func filterByError[T any](in []options.O[T], err error) []options.O[T] {
	result := make([]options.O[T], 0, len(in))
	for _, o := range in {
		if o.Err != nil && errors.Is(o.Err, err) {
			continue
		}
		result = append(result, o)
	}
	return result
}

// PoolPrice holds the informations of a certain Pool together with its total
// price.
type PoolPrice struct {
	Pool       LiquidityPool `json:"pool"`
	TotalPrice sdktypes.Dec  `json:"total_price"`
}

func NewPoolPrice(ctx context.Context, pool LiquidityPool, denomPricer DenomPricer) (PoolPrice, error) {
	price, err := pool.Price(ctx, denomPricer)
	if err != nil {
		return PoolPrice{}, fmt.Errorf("getting price for pool: %w", err)
	}
	return PoolPrice{
		Pool:       pool,
		TotalPrice: price,
	}, nil
}

// LiquidityPool is a generic representation of a liquidity pool. A pool has a
// price (in USD) that represents the total value of assets present inside the
// pool.
type LiquidityPool interface {
	Price(ctx context.Context, denomPricer DenomPricer) (sdktypes.Dec, error)
}

// MultiCoinPool represents a liquidity pool composed by one or more coins
// (denoms).
type MultiCoinPool struct {
	coins []sdktypes.Coin
}

var _ LiquidityPool = (*MultiCoinPool)(nil)

func NewMultiCoinPool() *MultiCoinPool { return &MultiCoinPool{} }

func (p *MultiCoinPool) AddCoin(coin sdktypes.Coin) {
	p.coins = append(p.coins, coin)
}

func (p *MultiCoinPool) Price(ctx context.Context, denomPricer DenomPricer) (sdktypes.Dec, error) {
	poolTVL := sdktypes.NewDec(0)
	for _, coin := range p.coins {
		denomPrice, err := denomPricer.DenomPrice(ctx, osmosisChainName, coin.Denom)
		if err != nil {
			return sdktypes.Dec{}, fmt.Errorf("getting denom price: %w", err)
		}
		poolTVL = poolTVL.Add(denomPrice.MulInt(coin.Amount))
	}

	return poolTVL, nil
}
