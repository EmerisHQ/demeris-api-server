package usecase

import (
	"context"
	"fmt"

	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	cretypes "github.com/crescent-network/crescent/x/liquidity/types"
	"google.golang.org/grpc"
)

type CrescentGrpcClient struct {
	URL string
}

var _ CrescentClient = &CrescentGrpcClient{}

func NewCrescentGrpcClient(url string) *CrescentGrpcClient {
	return &CrescentGrpcClient{
		URL: url,
	}
}

func (c *CrescentGrpcClient) Pools(ctx context.Context) ([]cretypes.PoolResponse, error) {
	// init gRPC osmosis client
	// TODO: when we have a service mesh reuse a single connection instead of
	// opening a new one every time
	grpcConn, err := grpc.Dial(c.URL, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("cannot connect to crescent, %w", err)
	}
	defer func() {
		_ = grpcConn.Close()
	}()

	creq := cretypes.NewQueryClient(grpcConn)

	var (
		pools   []cretypes.PoolResponse
		nextKey []byte
	)
	for {
		res, err := creq.Pools(ctx, &cretypes.QueryPoolsRequest{
			Pagination: &sdkquery.PageRequest{
				Key: nextKey,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("cannot get pools, %w", err)
		}
		pools = append(pools, res.Pools...)

		if len(res.Pagination.NextKey) == 0 {
			break
		}
		nextKey = res.Pagination.NextKey
	}

	return pools, nil
}
