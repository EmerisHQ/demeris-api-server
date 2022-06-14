package usecase

import (
	"context"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	cretypes "github.com/crescent-network/crescent/x/liquidity/types"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	gammbalancer "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

//go:generate mockgen -package usecase_test -source ports.go -destination ports_mocks_test.go

type SDKServiceClients interface {
	GetSDKServiceClient(version string) (sdkutilities.Service, error)
}

type SDKServiceClient interface {
	AccountNumbers(context.Context, *sdkutilities.AccountNumbersPayload) (res *sdkutilities.AccountNumbers2, err error)
	Supply(context.Context, *sdkutilities.SupplyPayload) (res *sdkutilities.Supply2, err error)
	SupplyDenom(context.Context, *sdkutilities.SupplyDenomPayload) (res *sdkutilities.Supply2, err error)
	QueryTx(context.Context, *sdkutilities.QueryTxPayload) (res []byte, err error)
	BroadcastTx(context.Context, *sdkutilities.BroadcastTxPayload) (res *sdkutilities.TransactionResult, err error)
	TxMetadata(context.Context, *sdkutilities.TxMetadataPayload) (res *sdkutilities.TxMessagesMetadata, err error)
	Block(context.Context, *sdkutilities.BlockPayload) (res *sdkutilities.BlockData, err error)
	LiquidityParams(context.Context, *sdkutilities.LiquidityParamsPayload) (res *sdkutilities.LiquidityParams2, err error)
	LiquidityPools(context.Context, *sdkutilities.LiquidityPoolsPayload) (res *sdkutilities.LiquidityPools2, err error)
	MintInflation(context.Context, *sdkutilities.MintInflationPayload) (res *sdkutilities.MintInflation2, err error)
	MintParams(context.Context, *sdkutilities.MintParamsPayload) (res *sdkutilities.MintParams2, err error)
	MintAnnualProvision(context.Context, *sdkutilities.MintAnnualProvisionPayload) (res *sdkutilities.MintAnnualProvision2, err error)
	MintEpochProvisions(context.Context, *sdkutilities.MintEpochProvisionsPayload) (res *sdkutilities.MintEpochProvisions2, err error)
	DelegatorRewards(context.Context, *sdkutilities.DelegatorRewardsPayload) (res *sdkutilities.DelegatorRewards2, err error)
	EstimateFees(context.Context, *sdkutilities.EstimateFeesPayload) (res *sdkutilities.Simulation, err error)
	StakingParams(context.Context, *sdkutilities.StakingParamsPayload) (res *sdkutilities.StakingParams2, err error)
	StakingPool(context.Context, *sdkutilities.StakingPoolPayload) (res *sdkutilities.StakingPool2, err error)
	EmoneyInflation(context.Context, *sdkutilities.EmoneyInflationPayload) (res *sdkutilities.EmoneyInflation2, err error)
	BudgetParams(context.Context, *sdkutilities.BudgetParamsPayload) (res *sdkutilities.BudgetParams2, err error)
	DistributionParams(context.Context, *sdkutilities.DistributionParamsPayload) (res *sdkutilities.DistributionParams2, err error)
}

type OsmosisClient interface {
	Pools(ctx context.Context) ([]gammbalancer.Pool, error)
}

type CrescentClient interface {
	Pools(ctx context.Context) ([]cretypes.PoolResponse, error)
}

// DenomPrices returns the USD price of a denom of the given chain.
type DenomPricer interface {
	DenomPrice(ctx context.Context, chainName, denom string) (sdktypes.Dec, error)
}
