package sdkservice_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/emerishq/demeris-api-server/sdkservice"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"github.com/stretchr/testify/require"
)

func TestSDKServiceURL(t *testing.T) {
	client, err := sdkservice.Client("45")
	require.NoError(t, err)
	port := 9090
	res, err := client.Block(context.Background(), &sdkutilities.BlockPayload{
		ChainName: "cosmos-hub",
		Port:      &port,
		Height:    1000000,
	})
	var data map[interface{}]interface{}
	require.NoError(t, json.Unmarshal(res.Block, &data))

	fmt.Println(data)
}
