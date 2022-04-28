package sdkservice

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/lib/apierrors"
	sdkserviceclient "github.com/emerishq/sdk-service-meta/gen/grpc/sdk_utilities/client"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"google.golang.org/grpc"
)

const (
	sdkServiceURLFmt = "sdk-service-v%s:9090"
)

var (
	availableVersions = []string{"42", "44"}

	// map of sdk versions to sdk service versions in case of any exceptions
	sdkExceptionMappings = map[string]string{
		"45": "44",
	}
)

func SdkServiceURL(version string) string {
	if v, ok := sdkExceptionMappings[version]; ok {
		version = v
	}
	return fmt.Sprintf(sdkServiceURLFmt, version)
}

type SDKServiceClients map[string]sdkutilities.Client

func (clients SDKServiceClients) GetSDKServiceClient(version string) (sdkutilities.Client, *apierrors.Error) {
	if v, ok := sdkExceptionMappings[version]; ok {
		version = v
	}

	client, ok := clients[version]
	if !ok {
		return client, apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve sdk-service for version %s", version),
			http.StatusBadRequest,
		)
	}

	return client, nil
}

func InitializeClients() (SDKServiceClients, error) {
	clients := SDKServiceClients{}
	for _, version := range availableVersions {
		client, err := Client(version)
		if err != nil {
			return clients, err
		}
		clients[version] = client
	}

	return clients, nil
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
		SupplyDenomEndpoint:         client.SupplyDenom(),
		QueryTxEndpoint:             client.QueryTx(),
		BroadcastTxEndpoint:         client.BroadcastTx(),
		TxMetadataEndpoint:          client.TxMetadata(),
		BlockEndpoint:               client.Block(),
		LiquidityParamsEndpoint:     client.LiquidityParams(),
		LiquidityPoolsEndpoint:      client.LiquidityPools(),
		MintInflationEndpoint:       client.MintInflation(),
		MintParamsEndpoint:          client.MintParams(),
		MintAnnualProvisionEndpoint: client.MintAnnualProvision(),
		MintEpochProvisionsEndpoint: client.MintEpochProvisions(),
		DelegatorRewardsEndpoint:    client.DelegatorRewards(),
		EstimateFeesEndpoint:        client.EstimateFees(),
		StakingParamsEndpoint:       client.StakingParams(),
		StakingPoolEndpoint:         client.StakingPool(),
		EmoneyInflationEndpoint:     client.EmoneyInflation(),
	}

	return cc, nil
}
