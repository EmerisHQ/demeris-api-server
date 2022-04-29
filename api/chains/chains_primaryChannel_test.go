package chains_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/stretchr/testify/require"
)

func TestGetPrimaryChannelWithCounterparty(t *testing.T) {

	require.NoError(t, testingCtx.CnsDB.AddChain(utils.ChainWithPublicEndpoints))

	body, status := getRespBodyAndStatus(t, "/chain/chain2/primary_channel/key", testingCtx.Cfg.ListenAddr)

	actual := getValFromByteArr(t, body, "primary_channel")

	expected := map[string]interface{}{
		"channel_name": "value",
		"counterparty": "key",
	}
	require.Equal(t, expected, actual)
	require.Equal(t, 200, status)

	body, status = getRespBodyAndStatus(t, "/chain/chain2/primary_channel/cosmos-hub", testingCtx.Cfg.ListenAddr)
	actual = getValFromByteArr(t, body, "cause")
	require.Equal(t, "cannot retrieve primary channel between chain2 and cosmos-hub", actual)
	require.Equal(t, 400, status)

	body, status = getRespBodyAndStatus(t, "/chain/osmosis/primary_channel/terra", testingCtx.Cfg.ListenAddr)
	actual = getValFromByteArr(t, body, "cause")
	require.Equal(t, "cannot retrieve chain with name osmosis", actual)
	require.Equal(t, 400, status)
}

func getValFromByteArr(t *testing.T, b []byte, key string) interface{} {
	t.Helper()

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &data))

	return data[key]
}

func getRespBodyAndStatus(t *testing.T, endPoint string, listenAddr string) ([]byte, int) {

	res, err := http.Get(fmt.Sprintf("http://%s%s", listenAddr, endPoint))
	require.NoError(t, err)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	return body, res.StatusCode
}
