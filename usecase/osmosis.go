package usecase

import (
	"context"
	"fmt"

	sdkcodec "github.com/cosmos/cosmos-sdk/codec/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	gammbalancer "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gamm "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"google.golang.org/grpc"
)

type OsmosisGrpcClient struct {
	URL string
}

var _ OsmosisClient = &OsmosisGrpcClient{}

func NewOsmosisGrpcClient(url string) *OsmosisGrpcClient {
	return &OsmosisGrpcClient{
		URL: url,
	}
}

func (c *OsmosisGrpcClient) Pools(ctx context.Context) ([]gammbalancer.Pool, error) {
	// init gRPC osmosis client
	// TODO: when we have a service mesh reuse a single connection instead of
	// opening a new one every time
	grpcConn, err := grpc.Dial(c.URL, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = grpcConn.Close()
	}()
	gq := gamm.NewQueryClient(grpcConn)

	// list pools (paginated)
	var (
		untypedPools []*sdkcodec.Any
		nextKey      []byte
	)
	for {
		res, err := gq.Pools(ctx, &gamm.QueryPoolsRequest{
			Pagination: &sdkquery.PageRequest{
				Key: nextKey,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("cannot get pools: %w", err)
		}
		untypedPools = append(untypedPools, res.Pools...)

		if len(res.Pagination.NextKey) == 0 {
			break
		}
		nextKey = res.Pagination.NextKey
	}

	// convert pools to gammbalancer.Pool
	pools := make([]gammbalancer.Pool, 0, len(untypedPools))
	for _, p := range untypedPools {
		var pool gammbalancer.Pool
		err := pool.Unmarshal(p.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal pool: %w", err)
		}
		pools = append(pools, pool)
	}

	return pools, nil
}
