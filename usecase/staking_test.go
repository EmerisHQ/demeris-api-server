package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-backend-models/cns"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStakingAPR(t *testing.T) {
	var (
		ctx                = context.Background()
		genericErr         = errors.New("oups")
		stakingPoolBytes   = []byte(`{"pool":{"bonded_tokens":"183301421577182"}}`)
		stakingParamsBytes = []byte(`{"params":{"bond_denom":"lamb"}}`)
		mintInflationBytes = []byte(`{"inflation":"0.112331651975797806"}`)
		budgetParamsBytes  = []byte(`{"params":{"budgets":[
      {"name":"budget-ecosystem-incentive","rate":"0.662500000000000000"},
      {"name":"xxx","rate":"1"},
      {"name":"budget-dev-team","rate":"0.250000000000000000"}
    ]}}`)
		distributionParamsBytes = []byte(`{"params":{
      "community_tax":"0.285714285700000000"
    }}`)
		mintParamsEmptyBytes = []byte(`{"params":{}}`)
		mintParamsBytes      = []byte(`{"params":{
      "inflation_schedules":[{
        "start_time":"2022-04-13T00:00:00Z",
        "end_time":"2122-04-13T00:00:00Z",
        "amount":"108700000000000"
      }]
    }}`)
	)
	tests := []struct {
		name          string
		chain         cns.Chain
		expectedError *apierrors.Error
		expectedAPR   string

		setup func(mocks)
	}{
		{
			name: "fail: sdkClient.StakingPool returns an error",
			chain: cns.Chain{
				ChainName: "lambda",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve staking pool from sdk-service",
				http.StatusBadRequest),

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "lambda",
				}).Return(nil, genericErr)
			},
		},
		{
			name: "fail: sdkClient.StakingParams returns an error",
			chain: cns.Chain{
				ChainName: "lambda",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve staking params from sdk-service",
				http.StatusBadRequest),

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "lambda",
				}).Return(nil, genericErr)
			},
		},
		{
			name: "fail: sdkClient.SupplyDenom returns an error",
			chain: cns.Chain{
				ChainName: "lambda",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve supply denom from sdk-service",
				http.StatusBadRequest),

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "lambda",
					Denom:     &([]string{"lamb"})[0],
				}).Return(nil, genericErr)
			},
		},
		{
			name: "fail: sdkClient.SupplyDenom returns multiple coins",
			chain: cns.Chain{
				ChainName: "lambda",
			},
			expectedError: apierrors.New("chains",
				"expected 1 denom for chain: lambda - denom: lamb, found 2",
				http.StatusBadRequest),

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "lambda",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "100000"},
						{Denom: "blamb", Amount: "1000000"},
					},
				}, nil)
			},
		},
		{
			name: "fail: sdkClient.MintInflation returns an error",
			chain: cns.Chain{
				ChainName: "lambda",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve inflation from sdk-service",
				http.StatusBadRequest),

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "lambda",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "296346737551905"},
					},
				}, nil)
				m.sdkClient.EXPECT().MintInflation(ctx, &sdkutilities.MintInflationPayload{
					ChainName: "lambda",
				}).Return(nil, genericErr)
			},
		},
		{
			name: "ok: lambda chain",
			chain: cns.Chain{
				ChainName: "lambda",
			},
			expectedAPR: "0.000000625591971500",

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "lambda",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "296346737551905"},
					},
				}, nil)
				m.sdkClient.EXPECT().MintInflation(ctx, &sdkutilities.MintInflationPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.MintInflation2{
					MintInflation: mintInflationBytes,
				}, nil)
			},
		},
		{
			name: "ok: osmosis chain",
			chain: cns.Chain{
				ChainName: "osmosis",
			},
			expectedAPR: "0.000000156397992900",

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "osmosis",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "296346737551905"},
					},
				}, nil)
				m.sdkClient.EXPECT().MintInflation(ctx, &sdkutilities.MintInflationPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.MintInflation2{
					MintInflation: mintInflationBytes,
				}, nil)
			},
		},
		{
			name: "ok: crescent chain, no inflation found",
			chain: cns.Chain{
				ChainName: "crescent",
			},
			expectedAPR: "0.000000000000000000",

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().BudgetParams(ctx, &sdkutilities.BudgetParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.BudgetParams2{
					BudgetParams: budgetParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().DistributionParams(ctx, &sdkutilities.DistributionParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.DistributionParams2{
					DistributionParams: distributionParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().MintParams(ctx, &sdkutilities.MintParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.MintParams2{
					MintParams: mintParamsEmptyBytes,
				}, nil)
			},
		},
		{
			name: "ok: crescent chain, inflation found",
			chain: cns.Chain{
				ChainName: "crescent",
			},
			expectedAPR: "4.324048228917309007",

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().BudgetParams(ctx, &sdkutilities.BudgetParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.BudgetParams2{
					BudgetParams: budgetParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().DistributionParams(ctx, &sdkutilities.DistributionParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.DistributionParams2{
					DistributionParams: distributionParamsBytes,
				}, nil)
				m.sdkClient.EXPECT().MintParams(ctx, &sdkutilities.MintParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.MintParams2{
					MintParams: mintParamsBytes,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			app := newApp(t, tt.setup)

			apr, err := app.StakingAPR(ctx, tt.chain)

			if tt.expectedError != nil {
				require.Equal(tt.expectedError, err)
				return
			}
			require.NoError(err)
			assert.Equal(tt.expectedAPR, apr.String())
		})
	}
}
