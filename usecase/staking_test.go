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
		ctx        = context.Background()
		genericErr = errors.New("oups")
		denomAtom  = "uatom"
		denomOsmo  = "uosmo"
	)
	tests := []struct {
		name          string
		chain         cns.Chain
		expectedError *apierrors.Error
		expectedAPR   string

		setup func(mockeds)
	}{
		{
			name: "fail: sdkClient.StakingPool returns an error",
			chain: cns.Chain{
				ChainName: "cosmos-hub",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve staking pool from sdk-service",
				http.StatusBadRequest),

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "cosmos-hub",
				}).Return(nil, genericErr)
			},
		},
		{
			name: "fail: sdkClient.StakingParams returns an error",
			chain: cns.Chain{
				ChainName: "cosmos-hub",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve staking params from sdk-service",
				http.StatusBadRequest),

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"183301421577182"}}`),
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "cosmos-hub",
				}).Return(nil, genericErr)
			},
		},
		{
			name: "fail: sdkClient.SupplyDenom returns an error",
			chain: cns.Chain{
				ChainName: "cosmos-hub",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve supply denom from sdk-service",
				http.StatusBadRequest),

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"183301421577182"}}`),
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: []byte(`{"params":{"bond_denom":"uatom"}}`),
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "cosmos-hub",
					Denom:     &denomAtom,
				}).Return(nil, genericErr)
			},
		},
		{
			name: "fail: sdkClient.SupplyDenom returns multiple coins",
			chain: cns.Chain{
				ChainName: "cosmos-hub",
			},
			expectedError: apierrors.New("chains",
				"expected 1 denom for chain: cosmos-hub - denom: uatom, found 2",
				http.StatusBadRequest),

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"183301421577182"}}`),
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: []byte(`{"params":{"bond_denom":"uatom"}}`),
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "cosmos-hub",
					Denom:     &denomAtom,
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "atom", Amount: "100000uatom"},
						{Denom: "batom", Amount: "1000000ubatom"},
					},
				}, nil)
			},
		},
		{
			name: "fail: sdkClient.MintInflation returns an error",
			chain: cns.Chain{
				ChainName: "cosmos-hub",
			},
			expectedError: apierrors.Wrap(genericErr, "chains",
				"cannot retrieve inflation from sdk-service",
				http.StatusBadRequest),

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"183301421577182"}}`),
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: []byte(`{"params":{"bond_denom":"uatom"}}`),
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "cosmos-hub",
					Denom:     &denomAtom,
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "atom", Amount: "296346737551905uatom"},
					},
				}, nil)
				m.sdkClient.EXPECT().MintInflation(ctx, &sdkutilities.MintInflationPayload{
					ChainName: "cosmos-hub",
				}).Return(nil, genericErr)
			},
		},
		{
			name: "ok: cosmos-hub chain",
			chain: cns.Chain{
				ChainName: "cosmos-hub",
			},
			expectedAPR: "18.160862201947935500",

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"183301421577182"}}`),
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: []byte(`{"params":{"bond_denom":"uatom"}}`),
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "cosmos-hub",
					Denom:     &denomAtom,
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "atom", Amount: "296346737551905uatom"},
					},
				}, nil)
				m.sdkClient.EXPECT().MintInflation(ctx, &sdkutilities.MintInflationPayload{
					ChainName: "cosmos-hub",
				}).Return(&sdkutilities.MintInflation2{
					MintInflation: []byte(`{"inflation":"0.112331651975797806"}`),
				}, nil)
			},
		},
		{
			name: "ok: osmosis chain",
			chain: cns.Chain{
				ChainName: "osmosis",
			},
			expectedAPR: "61.996006578275985200",

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"120975533972991"}}`),
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: []byte(`{"params":{"bond_denom":"uosmo"}}`),
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "osmosis",
					Denom:     &denomOsmo,
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "osmo", Amount: "377806829582915uosmo"},
					},
				}, nil)
				m.sdkClient.EXPECT().MintInflation(ctx, &sdkutilities.MintInflationPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.MintInflation2{
					MintInflation: []byte(`{"inflation":"0.794056582648834299"}`),
				}, nil)
			},
		},
		{
			name: "ok: crescent chain, no inflation found",
			chain: cns.Chain{
				ChainName: "crescent",
			},
			expectedAPR: "0.000000000000000000",

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"17907124553766"}}`),
				}, nil)
				m.sdkClient.EXPECT().BudgetParams(ctx, &sdkutilities.BudgetParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.BudgetParams2{
					BudgetParams: []byte(`{"params":{"budgets":[
              {"name":"budget-ecosystem-incentive","rate":"0.662500000000000000"},
              {"name":"xxx","rate":"1"},
              {"name":"budget-dev-team","rate":"0.250000000000000000"}
            ]}}`),
				}, nil)
				m.sdkClient.EXPECT().DistributionParams(ctx, &sdkutilities.DistributionParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.DistributionParams2{
					DistributionParams: []byte(`{"params":{
            "community_tax":"0.285714285700000000"
          }}`),
				}, nil)
				m.sdkClient.EXPECT().MintParams(ctx, &sdkutilities.MintParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.MintParams2{
					MintParams: []byte(`{"params":{}}`),
				}, nil)
			},
		},
		{
			name: "ok: crescent chain, inflation found",
			chain: cns.Chain{
				ChainName: "crescent",
			},
			expectedAPR: "37.938810219014751900",

			setup: func(m mockeds) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: []byte(`{"pool":{"bonded_tokens":"17907124553766"}}`),
				}, nil)
				m.sdkClient.EXPECT().BudgetParams(ctx, &sdkutilities.BudgetParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.BudgetParams2{
					BudgetParams: []byte(`{"params":{"budgets":[
              {"name":"budget-ecosystem-incentive","rate":"0.662500000000000000"},
              {"name":"xxx","rate":"1"},
              {"name":"budget-dev-team","rate":"0.250000000000000000"}
            ]}}`),
				}, nil)
				m.sdkClient.EXPECT().DistributionParams(ctx, &sdkutilities.DistributionParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.DistributionParams2{
					DistributionParams: []byte(`{"params":{
            "community_tax":"0.285714285700000000"
          }}`),
				}, nil)
				m.sdkClient.EXPECT().MintParams(ctx, &sdkutilities.MintParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.MintParams2{
					MintParams: []byte(`{"params":{
              "inflation_schedules":[
              {
                "start_time": "2022-04-13T00:00:00Z",
                "end_time": "2023-04-13T00:00:00Z",
                "amount": "108700000000000"
              },
              {
                "start_time": "2023-04-13T00:00:00Z",
                "end_time": "2024-04-13T00:00:00Z",
                "amount": "216100000000000"
              }
              ]
            }}`),
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
