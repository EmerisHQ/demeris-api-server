package usecase

import (
	"context"
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	cretypes "github.com/crescent-network/crescent/x/liquidity/types"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

type CrescentPoolsStrategy struct {
	client            CrescentClient
	sdkServiceClients SDKServiceClients
}

var _ PoolsStrategy = (*CrescentPoolsStrategy)(nil)

func NewCrescentPoolsStrategy(
	client CrescentClient,
	sdkServiceClients SDKServiceClients,
) *CrescentPoolsStrategy {
	return &CrescentPoolsStrategy{
		client:            client,
		sdkServiceClients: sdkServiceClients,
	}
}

func (s *CrescentPoolsStrategy) ID() StrategyID { return crescentChainName }

func (s *CrescentPoolsStrategy) Pools(ctx context.Context) ([]LiquidityPool, error) {
	pools, err := s.client.Pools(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing pools: %w", err)
	}

	// list supplies (paginated)
	v := "44" // crescent version, TODO: fetch this from cns/database
	sdkService, err := s.sdkServiceClients.GetSDKServiceClient(v)
	if err != nil {
		return nil, fmt.Errorf("cannot get SDK service client, %w", err)
	}
	var (
		nextSupplyKey *string
		supplies      = make(map[string]string)
	)
	for {
		res, err := sdkService.Supply(ctx, &sdkutilities.SupplyPayload{
			ChainName:     "crescent",
			PaginationKey: nextSupplyKey,
		})
		if err != nil {
			return nil, fmt.Errorf("cannot get pools, %w", err)
		}
		for _, s := range res.Coins {
			supplies[s.Denom] = s.Amount
		}

		if res.Pagination.NextKey == nil || len(*res.Pagination.NextKey) == 0 {
			break
		}
		nextSupplyKey = res.Pagination.NextKey
	}

	// get prices for each pool
	res := make([]LiquidityPool, 0, len(pools))
	for _, p := range pools {
		res = append(res, NewCrescentPool(p, supplies))
	}

	return res, nil
}

type CrescentPool struct {
	*MultiCoinPool

	Denom       string `json:"denom"`
	TotalSupply string `json:"total_supply"`
}

var _ LiquidityPool = CrescentPool{}

func NewCrescentPool(p cretypes.PoolResponse, supplies map[string]string) *CrescentPool {
	totalSupply, found := supplies[p.PoolCoinDenom]
	if !found {
		totalSupply = "0"
	}

	pool := &CrescentPool{
		MultiCoinPool: NewMultiCoinPool(),
		Denom:         p.PoolCoinDenom,
		TotalSupply:   totalSupply,
	}

	for _, b := range p.Balances {
		pool.AddCoin(sdktypes.NewCoin(b.Denom, b.Amount))
	}

	return pool
}
