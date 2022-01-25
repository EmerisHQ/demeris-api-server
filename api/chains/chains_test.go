package chains_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/allinbits/demeris-api-server/api/config"
	"github.com/allinbits/demeris-api-server/api/database"
	"github.com/allinbits/demeris-api-server/api/router"
	"github.com/allinbits/demeris-backend-models/cns"
	cnsDB "github.com/allinbits/emeris-cns-server/cns/database"
	"github.com/allinbits/emeris-utils/logging"
	"github.com/allinbits/emeris-utils/store"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetPrimaryChannelWithCounterparty(t *testing.T) {
	r, client, instance, cfg, _, tDown := setup(t)
	defer tDown()

	require.NotNil(t, r)

	primaryChannels := cns.DbStringMap{
		"terra": "channel-1",
	}
	newChain := cns.Chain{
		ChainName:        "akash",
		Enabled:          true,
		PrimaryChannel:   primaryChannels,
		DemerisAddresses: pq.StringArray{},
		SupportedWallets: pq.StringArray{},
	}

	require.NoError(t, instance.AddChain(newChain))

	body := getRespBody(t, "/chain/akash/primary_channel/terra", client, cfg.ListenAddr)

	actual := getValFromByteArr(t, body, "primary_channel")

	expected := map[string]interface{}{
		"channel_name": "channel-1",
		"counterparty": "terra",
	}
	require.Equal(t, expected, actual)

	body = getRespBody(t, "/chain/akash/primary_channel/cosmos-hub", client, cfg.ListenAddr)
	actual = getValFromByteArr(t, body, "cause")
	require.Equal(t, "cannot retrieve primary channel between akash and cosmos-hub", actual)

	body = getRespBody(t, "/chain/osmosis/primary_channel/terra", client, cfg.ListenAddr)
	actual = getValFromByteArr(t, body, "cause")
	require.Equal(t, "cannot retrieve chain with name osmosis", actual)
}

func setup(t *testing.T) (router.Router, http.Client, cnsDB.Instance, config.Config, zap.SugaredLogger, func()) {
	t.Helper()
	ts, err := testserver.NewTestServer()
	require.NoError(t, err)

	require.NoError(t, ts.WaitForInit())

	connStr := ts.PGURL().String()
	cnsInstance, err := cnsDB.New(connStr)
	require.NoError(t, err)

	cfg := &config.Config{
		DatabaseConnectionURL: connStr,
		ListenAddr:            "127.0.0.1:9090",
		RedisAddr:             "127.0.0.1:6379",
		KubernetesNamespace:   "emeris",
		Debug:                 true,
	}

	db, err := database.Init(cfg)
	require.NoError(t, err)

	s, err := store.NewClient(cfg.RedisAddr)
	require.NoError(t, err)

	l := logging.New(logging.LoggingConfig{
		Debug: cfg.Debug,
	})

	r := router.New(db, l, s, nil, "emeris", nil, cfg.Debug)

	go r.Serve(cfg.ListenAddr)
	time.Sleep(2 * time.Second)

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	return *r, client, *cnsInstance, *cfg, *l, func() { ts.Stop() }
}

func getValFromByteArr(t *testing.T, b []byte, key string) interface{} {
	t.Helper()

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &data))

	return data[key]
}

func getRespBody(t *testing.T, endPoint string, client http.Client, listenAddr string) []byte {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s%s", listenAddr, endPoint), nil)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	return body
}
