package staking

import (
	"context"

	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

//go:generate mockgen -source ports.go -destination ports_mocks_test.go -package staking_test

type SDKClient interface {
	StakingPool(context.Context, *sdkutilities.StakingPoolPayload) (*sdkutilities.StakingPool2, error)
	StakingParams(context.Context, *sdkutilities.StakingParamsPayload) (*sdkutilities.StakingParams2, error)
	SupplyDenom(context.Context, *sdkutilities.SupplyDenomPayload) (*sdkutilities.Supply2, error)
	MintInflation(context.Context, *sdkutilities.MintInflationPayload) (*sdkutilities.MintInflation2, error)
}
