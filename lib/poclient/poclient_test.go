package poclient_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/emerishq/demeris-api-server/lib/poclient"
	cnsDb "github.com/emerishq/emeris-cns-server/cns/database"
	"github.com/emerishq/emeris-price-oracle/price-oracle/config"
	"github.com/emerishq/emeris-price-oracle/price-oracle/rest"
	"github.com/emerishq/emeris-price-oracle/price-oracle/sql"
	"github.com/emerishq/emeris-price-oracle/price-oracle/store"
	potypes "github.com/emerishq/emeris-price-oracle/price-oracle/types"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/stretchr/testify/require"
)

var testToken = potypes.TokenPriceAndSupply{
	Symbol: strings.ToUpper(utils.ChainWithPublicEndpoints.Denoms[0].Name + potypes.USDT),
	Price:  50,
	Supply: 100000,
}

var testFiat = potypes.FiatPrice{
	Symbol: strings.ToUpper(potypes.USD + utils.ChainWithPublicEndpoints.Denoms[0].Name),
	Price:  2,
}

func TestGetPrice(t *testing.T) {
	testServer, listenAddr := setup(t)
	defer tearDown(testServer)

	poURL := fmt.Sprintf("http://%s", listenAddr)

	tests := []struct {
		name      string
		poBaseURL string
		symbol    string
		success   bool
	}{
		{
			"invalid price oracle base url",
			"http://invalid.com",
			"symbol",
			false,
		},
		{
			"invalid symbol",
			poURL,
			"symbol",
			false,
		},
		{
			"valid token symbol",
			poURL,
			testToken.Symbol,
			true,
		},
		{
			"valid fiat symbol",
			poURL,
			testFiat.Symbol,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := poclient.NewPOClient(tt.poBaseURL)
			price, err := client.GetPrice(context.Background(), tt.symbol)
			if tt.success {
				require.NoError(t, err)
				require.NotEmpty(t, price)
				require.Positive(t, price.Price)
			} else {
				require.Error(t, err)
				require.Empty(t, price)
			}
		})
	}
}

func setup(t *testing.T) (testserver.TestServer, string) {
	t.Helper()
	ts, err := testserver.NewTestServer()
	require.NoError(t, err)
	require.NoError(t, ts.WaitForInit())

	c := &config.Config{
		LogPath:               "",
		Debug:                 true,
		DatabaseConnectionURL: ts.PGURL().String(),
		Interval:              "10s",
		WhitelistedFiats:      []string{strings.ToUpper(utils.ChainWithPublicEndpoints.Denoms[0].Name)},
		ListenAddr:            "FILLME",
		MaxAssetsReq:          10,
		HttpClientTimeout:     1 * time.Second,
	}

	cns, err := cnsDb.New(c.DatabaseConnectionURL)
	require.NoError(t, ts.WaitForInit())

	require.NoError(t, cns.AddChain(utils.ChainWithPublicEndpoints))

	l := logging.New(logging.LoggingConfig{
		LogPath: "",
		Debug:   c.Debug,
	})

	db, err := sql.NewDB(c.DatabaseConnectionURL)
	require.NoError(t, err)

	ctx := context.Background()
	storeHandler, err := store.NewStoreHandler(
		store.WithDB(ctx, db),
		store.WithConfig(c),
		store.WithLogger(l),
		store.WithSpotPriceCache(nil),
		store.WithChartDataCache(nil, time.Minute*5),
	)
	require.NoError(t, err)

	err = db.Init(ctx)
	require.NoError(t, err)

	err = db.UpsertPrice(ctx, store.TokensStore, testToken.Price, testToken.Symbol)
	require.NoError(t, err)

	err = db.UpsertTokenSupply(ctx, store.CoingeckoSupplyStore, testToken.Symbol, testToken.Supply)
	require.NoError(t, err)

	err = db.UpsertPrice(ctx, store.FiatsStore, testFiat.Price, testFiat.Symbol)
	require.NoError(t, err)

	port, err := utils.GetFreePort()
	require.NoError(t, err)
	c.ListenAddr = "127.0.0.1:" + port

	r := rest.NewServer(storeHandler, l, c)

	ch := make(chan struct{})
	go func() {
		close(ch)
		err := r.Serve(c.ListenAddr)
		require.NoError(t, err)
	}()
	<-ch // Wait for the goroutine to start

	return ts, c.ListenAddr
}

func tearDown(ts testserver.TestServer) {
	ts.Stop()
}
