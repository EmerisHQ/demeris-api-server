package sdkservice

import (
	//sdkservicetypes "github.com/allinbits/sdk-service-meta"

	"fmt"

	sdkserviceclient "github.com/allinbits/sdk-service-meta/gen/grpc/sdk_utilities/client"
	sdkutilities "github.com/allinbits/sdk-service-meta/gen/sdk_utilities"
	"google.golang.org/grpc"
)

const sdkServiceURLFmt = "sdk-service-v%s"

func sdkServiceURL(version string) string {
	return fmt.Sprintf(sdkServiceURLFmt, version)
}

func Client(sdkVersion string) (sdkutilities.Client, error) {
	conn, err := grpc.Dial(sdkServiceURL(sdkVersion), grpc.WithInsecure())
	if err != nil {
		return sdkutilities.Client{}, fmt.Errorf("cannot connect to endpoint %s: %w", sdkServiceURL(sdkVersion), err)
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
	}

	return cc, nil
}
