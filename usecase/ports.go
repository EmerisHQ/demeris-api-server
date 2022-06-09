package usecase

import (
	"context"

	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

//go:generate mockgen -package usecase_test -source ports.go -destination ports_mocks_test.go

type SDKServiceClients interface {
	GetSDKServiceClient(version string) (SDKServiceClient, error)
}

type SDKServiceClient interface {
	AccountNumbers(context.Context, *sdkutilities.AccountNumbersPayload) (res *sdkutilities.AccountNumbers2, err error)
	SupplyDenom(context.Context, *sdkutilities.SupplyDenomPayload) (res *sdkutilities.Supply2, err error)
	MintInflation(context.Context, *sdkutilities.MintInflationPayload) (res *sdkutilities.MintInflation2, err error)
	MintParams(context.Context, *sdkutilities.MintParamsPayload) (res *sdkutilities.MintParams2, err error)
	StakingPool(context.Context, *sdkutilities.StakingPoolPayload) (res *sdkutilities.StakingPool2, err error)
	StakingParams(context.Context, *sdkutilities.StakingParamsPayload) (res *sdkutilities.StakingParams2, err error)
	BudgetParams(context.Context, *sdkutilities.BudgetParamsPayload) (res *sdkutilities.BudgetParams2, err error)
	DistributionParams(context.Context, *sdkutilities.DistributionParamsPayload) (res *sdkutilities.DistributionParams2, err error)
}
