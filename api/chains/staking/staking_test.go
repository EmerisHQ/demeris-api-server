package staking_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/emerishq/demeris-api-server/api/chains/staking"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-backend-models/cns"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mocks struct {
	t         *testing.T
	sdkClient *MockSDKClient
}

func newStaking(t *testing.T, setup func(mocks)) *staking.Staking {
	ctrl := gomock.NewController(t)
	m := mocks{
		t:         t,
		sdkClient: NewMockSDKClient(ctrl),
	}
	if setup != nil {
		setup(m)
	}
	return staking.New(m.sdkClient)
}

func TestStakingAPR(t *testing.T) {
	var (
		ctx                = context.Background()
		genericErr         = errors.New("oups")
		stakingPoolBytes   = []byte(`{"pool":{"bonded_tokens":"50000000"}}`)
		stakingParamBytes  = []byte(`{"params":{"bond_denom":"lamb"}}`)
		mintInflationBytes = []byte(`{"inflation":"10.0"}`)
	)
	tests := []struct {
		name          string
		chain         cns.Chain
		expectedError *apierrors.Error
		expectedAPR   float64

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
					StakingParams: stakingParamBytes,
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
					StakingParams: stakingParamBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "lambda",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "100000uatom"},
						{Denom: "blamb", Amount: "1000000ubatom"},
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
					StakingParams: stakingParamBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "lambda",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "1000000uatom"},
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
			expectedAPR: 20,

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "lambda",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "lambda",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "1000000uatom"},
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
			expectedAPR: 5,

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "osmosis",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "osmosis",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "1000000uatom"},
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
			name: "ok: crescent chain",
			chain: cns.Chain{
				ChainName: "crescent",
			},
			expectedAPR: 5,

			setup: func(m mocks) {
				m.sdkClient.EXPECT().StakingPool(ctx, &sdkutilities.StakingPoolPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.StakingPool2{
					StakingPool: stakingPoolBytes,
				}, nil)
				m.sdkClient.EXPECT().StakingParams(ctx, &sdkutilities.StakingParamsPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.StakingParams2{
					StakingParams: stakingParamBytes,
				}, nil)
				m.sdkClient.EXPECT().SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
					ChainName: "crescent",
					Denom:     &([]string{"lamb"})[0],
				}).Return(&sdkutilities.Supply2{
					Coins: []*sdkutilities.Coin{
						{Denom: "lamb", Amount: "1000000uatom"},
					},
				}, nil)
				m.sdkClient.EXPECT().MintInflation(ctx, &sdkutilities.MintInflationPayload{
					ChainName: "crescent",
				}).Return(&sdkutilities.MintInflation2{
					MintInflation: mintInflationBytes,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			s := newStaking(t, tt.setup)

			apr, err := s.APR(ctx, tt.chain)

			if tt.expectedError != nil {
				require.Equal(tt.expectedError, err)
				return
			}
			require.NoError(err)
			assert.Equal(tt.expectedAPR, apr)
		})
	}
}
