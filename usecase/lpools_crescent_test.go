package usecase_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	cretypes "github.com/crescent-network/crescent/x/liquidity/types"
	"github.com/emerishq/demeris-api-server/usecase"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCrescentPoolsStrategy_Pools(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		client            *MockCrescentClient
		sdkServiceClients *MockSDKServiceClients
	}
	tests := []struct {
		name     string
		supplies []*sdkutilities.Coin
		pools    []cretypes.PoolResponse
	}{
		{
			name: "ok: no pools",
			supplies: []*sdkutilities.Coin{
				{
					Denom:  "uatom",
					Amount: "10",
				},
			},
			pools: []cretypes.PoolResponse{},
		},
		{
			name: "ok: pool with two denoms in it",
			supplies: []*sdkutilities.Coin{
				{
					Denom:  "pool/1",
					Amount: "12341234",
				},
			},
			pools: []cretypes.PoolResponse{
				{
					Id: 1,
					Balances: []types.Coin{
						types.NewInt64Coin("uatom", 10),
						types.NewInt64Coin("ucre", 5),
					},
					PoolCoinDenom: "pool/1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := usecase.NewCrescentPoolsStrategy(makeArgs(t, ctx, tt.pools, tt.supplies))
			got, err := s.Pools(ctx)
			require.NoError(t, err)
			// TODO: this test table only checks that the len of returned
			// LiquidityPools matched the len of Crescent pools returned by the
			// Crecent client.
			// I wasn't able to write a simple but more accurate test (e.g.
			// checking that Amount and Denom are correct) because of the
			// MultiDenomPool inner struct.
			require.Len(t, got, len(tt.pools))
		})
	}
}

func makeArgs(t *testing.T, ctx context.Context, pools []cretypes.PoolResponse, supplies []*sdkutilities.Coin) (*MockCrescentClient, *MockSDKServiceClients) {
	ctrl := gomock.NewController(t)

	client := NewMockCrescentClient(ctrl)
	client.EXPECT().Pools(gomock.Any()).Return(pools, nil)

	sdkService := NewMockSDKServiceClient(ctrl)
	sdkService.EXPECT().Supply(ctx, gomock.Any()).Return(&sdkutilities.Supply2{
		Coins: supplies,
	}, nil)
	sdkServiceClients := NewMockSDKServiceClients(ctrl)
	sdkServiceClients.EXPECT().GetSDKServiceClient(gomock.Any()).Return(sdkService, nil)

	return client, sdkServiceClients
}
