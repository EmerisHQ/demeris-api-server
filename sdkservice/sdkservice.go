package sdkservice

import (
	"fmt"

	sdkserviceclient "github.com/emerishq/sdk-service-meta/gen/grpc/sdk_utilities/client"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"google.golang.org/grpc"
)

const (
	sdkServiceURLFmt = "sdk-service-v%s:9090"
)

// map of sdk versions to sdk service versions in case of any exceptions
var sdkExceptionMappings = map[string]string{
	"45": "44",
}

func SdkServiceURL(version string) string {
	if v, ok := sdkExceptionMappings[version]; ok {
		version = v
	}
	return fmt.Sprintf(sdkServiceURLFmt, version)
}

// Client returns a sdkutilities.Client for the given SDK version ready to be used.
func Client(sdkVersion string) (sdkutilities.Client, error) {
	conn, err := grpc.Dial(SdkServiceURL(sdkVersion), grpc.WithInsecure())
	if err != nil {
		return sdkutilities.Client{}, fmt.Errorf("cannot connect to endpoint %s: %w", SdkServiceURL(sdkVersion), err)
	}

	client := sdkserviceclient.NewClient(conn)

	cc := sdkutilities.Client{
		AccountNumbersEndpoint:      client.AccountNumbers(),
		SupplyEndpoint:              client.Supply(),
		QueryTxEndpoint:             client.QueryTx(),
		BroadcastTxEndpoint:         client.BroadcastTx(),
		TxMetadataEndpoint:          client.TxMetadata(),
		BlockEndpoint:               client.Block(),
		LiquidityParamsEndpoint:     client.LiquidityParams(),
		LiquidityPoolsEndpoint:      client.LiquidityPools(),
		MintInflationEndpoint:       client.MintInflation(),
		MintParamsEndpoint:          client.MintParams(),
		MintAnnualProvisionEndpoint: client.MintAnnualProvision(),
		EstimateFeesEndpoint:        client.EstimateFees(),
		DelegatorRewardsEndpoint:    client.DelegatorRewards(),
		StakingParamsEndpoint:       client.StakingParams(),
		StakingPoolEndpoint:         client.StakingPool(),
	}

	return cc, nil
}
