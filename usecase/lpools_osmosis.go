package usecase

import (
	"context"
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	gammbalancer "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

type OsmosisPoolsStrategy struct {
	client OsmosisClient
}

var _ PoolsStrategy = (*OsmosisPoolsStrategy)(nil)

func NewOsmosisPoolsStrategy(client OsmosisClient) *OsmosisPoolsStrategy {
	return &OsmosisPoolsStrategy{
		client: client,
	}
}

func (s *OsmosisPoolsStrategy) ID() StrategyID { return osmosisChainName }

func (s *OsmosisPoolsStrategy) Pools(ctx context.Context) ([]LiquidityPool, error) {
	pools, err := s.client.Pools(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting pools: %w", err)
	}

	// get prices for each pool
	res := make([]LiquidityPool, 0, len(pools))
	for _, p := range pools {
		res = append(res, NewOsmosisPool(p))
	}
	return res, nil
}

type OsmosisPool struct {
	*MultiCoinPool

	Denom       string       `json:"denom"`
	TotalSupply sdktypes.Int `json:"total_supply"`
}

var _ LiquidityPool = OsmosisPool{}

func NewOsmosisPool(p gammbalancer.Pool) *OsmosisPool {
	pool := &OsmosisPool{
		MultiCoinPool: NewMultiCoinPool(),
		Denom:         p.TotalShares.Denom,
		TotalSupply:   p.TotalShares.Amount,
	}
	for _, asset := range p.PoolAssets {
		pool.AddCoin(asset.Token)
	}
	return pool
}
