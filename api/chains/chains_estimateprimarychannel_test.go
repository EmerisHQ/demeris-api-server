package chains_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/emerishq/demeris-api-server/api/chains"
	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"github.com/lib/pq"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/stretchr/testify/require"
)

const (
	estimatePrimaryChannelsUrl = "http://%s/chains/primary_channels"
)

var tracelistenerData = utils.TracelistenerData{
	Denoms: []utils.DenomTrace{
		{
			Path:      "transfer/channel-0",
			BaseDenom: "uatom",
			Hash:      "27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			ChainName: "osmosis",
		},
		{
			Path:      "transfer/channel-141",
			BaseDenom: "uosmo",
			Hash:      "14F9BC3E44B8A9C1BE1FB08980FAB87034C9905EF17CF2F5008FC085218811CC",
			ChainName: "cosmos-hub",
		},
	},
	Channels: []utils.Channel{
		{
			ChannelID:        "channel-0",
			CounterChannelID: "channel-141",
			Port:             "transfer",
			State:            3,
			Hops:             []string{"connection-0"},
			ChainName:        "osmosis",
		},
		{
			ChannelID:        "channel-141",
			CounterChannelID: "channel-0",
			Port:             "transfer",
			State:            3,
			Hops:             []string{"connection-0"},
			ChainName:        "cosmos-hub",
		},
	},
	Connections: []utils.Connection{
		{
			ChainName:           "osmosis",
			ConnectionID:        "connection-0",
			ClientID:            "00-tendermint-69",
			State:               "STATE_OPEN",
			CounterConnectionID: "connection-0",
			CounterClientID:     "00-tendermint-69",
		},
		{
			ChainName:           "cosmos-hub",
			ConnectionID:        "connection-0",
			ClientID:            "00-tendermint-69",
			State:               "STATE_OPEN",
			CounterConnectionID: "connection-0",
			CounterClientID:     "00-tendermint-69",
		},
	},
	Clients: []utils.Client{
		{
			SourceChainName: "osmosis",
			DestChainID:     "cosmoshub-4",
			ClientID:        "00-tendermint-69",
			LatestHeight:    "42069",
			TrustingPeriod:  "9001",
		},
		{
			SourceChainName: "cosmos-hub",
			DestChainID:     "osmosis-1",
			ClientID:        "00-tendermint-69",
			LatestHeight:    "42069",
			TrustingPeriod:  "9001",
		},
	},
	BlockTimes: []utils.BlockTime{
		{
			ChainName: "osmosis",
			Time:      time.Now().UTC(),
		},
		{
			ChainName: "cosmos-hub",
			Time:      time.Now().UTC(),
		},
	},
}

var supplyDataOsmosis = []sdkutilities.Supply2{
	{
		Coins: []*sdkutilities.Coin{
			{
				Denom:  "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
				Amount: "100000000",
			},
		},
		Pagination: nil,
	},
}

var supplyDataCosmosHub = []sdkutilities.Supply2{
	{
		Coins: []*sdkutilities.Coin{
			{
				Denom:  "ibc/14F9BC3E44B8A9C1BE1FB08980FAB87034C9905EF17CF2F5008FC085218811CC",
				Amount: "100000000",
			},
		},
		Pagination: nil,
	},
}

func buildChainsEstimatePrimaryChannelResponse() chains.ChainsPrimaryChannelResponse {
	r := chains.ChainsPrimaryChannelResponse{
		FailureLogs: make(chains.FailLogs, 0),
	}

	r.Chains = make(map[string]map[string]chains.PrimaryChannelEstimation)
	r.Chains["cosmos-hub"] = make(map[string]chains.PrimaryChannelEstimation)
	r.Chains["osmosis"] = make(map[string]chains.PrimaryChannelEstimation)
	r.Chains["cosmos-hub"]["osmosis"] = chains.PrimaryChannelEstimation{
		CurrentPrimaryChannel:         "channel-141",
		EstimatedPrimaryChannel:       "channel-141",
		EstimatedPrimaryChannelDenom:  "ibc/14F9BC3E44B8A9C1BE1FB08980FAB87034C9905EF17CF2F5008FC085218811CC",
		EstimatedPrimaryChannelSupply: 100000000,
	}
	r.Chains["osmosis"]["cosmos-hub"] = chains.PrimaryChannelEstimation{
		CurrentPrimaryChannel:         "channel-0",
		EstimatedPrimaryChannel:       "channel-0",
		EstimatedPrimaryChannelDenom:  "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		EstimatedPrimaryChannelSupply: 100000000,
	}

	return r
}

func buildStoreDataSet() utils.StoreDataSet {
	d := make(utils.StoreDataSet, 0)

	bz, err := json.Marshal(supplyDataCosmosHub[0])
	if err != nil {
		panic(err)
	}

	d = append(d, utils.StoreData{
		Key:   "cosmos-hub-ibc/14F9BC3E44B8A9C1BE1FB08980FAB87034C9905EF17CF2F5008FC085218811CC",
		Value: bz,
	})

	bz, err = json.Marshal(supplyDataOsmosis[0])
	if err != nil {
		panic(err)
	}

	d = append(d, utils.StoreData{
		Key:   "osmosis-ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		Value: bz,
	})

	return d
}

func twochainz() []cns.Chain {
	relayerBalance := int64(42069)
	return []cns.Chain{
		{
			Enabled:        true,
			ChainName:      "cosmos-hub",
			Logo:           "http://logo.com",
			DisplayName:    "cosmos-hub",
			PrimaryChannel: map[string]string{"osmosis": "channel-141"},
			Denoms: []cns.Denom{
				{
					Name:        "uatom",
					DisplayName: "ATOM",
					Logo:        "http://logo.com",
					Precision:   8,
					Verified:    true,
					Stakable:    true,
					Ticker:      "DENOM1",
					PriceID:     "price_id_1",
					FeeToken:    true,
					GasPriceLevels: cns.GasPrice{
						Low:     0.2,
						Average: 0.3,
						High:    0.4,
					},
					FetchPrice:                  true,
					RelayerDenom:                true,
					MinimumThreshRelayerBalance: &relayerBalance,
				},
			},
			DemerisAddresses: []string{"12345"},
			GenesisHash:      "hash",
			NodeInfo: cns.NodeInfo{
				Endpoint: "http://endpoint",
				ChainID:  "cosmoshub-4",
				Bech32Config: cns.Bech32Config{
					MainPrefix:      "prefix",
					PrefixAccount:   "acc",
					PrefixValidator: "val",
					PrefixConsensus: "cons",
					PrefixPublic:    "pub",
					PrefixOperator:  "oper",
				},
			},
			ValidBlockThresh: cns.Threshold(30 * time.Minute),
			DerivationPath:   "m/44'/60'/0'/1",
			SupportedWallets: pq.StringArray([]string{"keplr"}),
			BlockExplorer:    "http://explorer.com",
		},
		{
			Enabled:        true,
			ChainName:      "osmosis",
			Logo:           "http://logo.com",
			DisplayName:    "osmosis",
			PrimaryChannel: map[string]string{"cosmos-hub": "channel-0"},
			Denoms: []cns.Denom{
				{
					Name:        "uosmo",
					DisplayName: "OSMO",
					Logo:        "http://logo.com",
					Precision:   8,
					Verified:    true,
					Stakable:    true,
					Ticker:      "DENOM1",
					PriceID:     "price_id_1",
					FeeToken:    true,
					GasPriceLevels: cns.GasPrice{
						Low:     0.2,
						Average: 0.3,
						High:    0.4,
					},
					FetchPrice:                  true,
					RelayerDenom:                true,
					MinimumThreshRelayerBalance: &relayerBalance,
				},
			},
			DemerisAddresses: []string{"12345"},
			GenesisHash:      "hash",
			NodeInfo: cns.NodeInfo{
				Endpoint: "http://endpoint",
				ChainID:  "osmosis-1",
				Bech32Config: cns.Bech32Config{
					MainPrefix:      "prefix",
					PrefixAccount:   "acc",
					PrefixValidator: "val",
					PrefixConsensus: "cons",
					PrefixPublic:    "pub",
					PrefixOperator:  "oper",
				},
			},
			ValidBlockThresh: cns.Threshold(30 * time.Minute),
			DerivationPath:   "m/44'/60'/0'/1",
			SupportedWallets: pq.StringArray([]string{"keplr"}),
			BlockExplorer:    "http://explorer.com",
		},
	}
}

func TestEstimatePrimaryChannels(t *testing.T) {

	storeData := buildStoreDataSet()
	expectedResponse := buildChainsEstimatePrimaryChannelResponse()

	tests := []struct {
		name              string
		chains            []cns.Chain
		tracelistenerData utils.TracelistenerData
		storeData         utils.StoreDataSet
		expectedHttpCode  int
		expectedResponse  chains.ChainsPrimaryChannelResponse
		success           bool
	}{
		{
			"Estimate primary channels - cosmos-hub and osmosis",
			twochainz(), // do something
			tracelistenerData,
			storeData,
			200,
			expectedResponse,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for _, chs := range tt.chains {
				err := testingCtx.CnsDB.AddChain(chs)
				require.NoError(t, err)
			}

			utils.RunTraceListnerMigrations(testingCtx, t)
			utils.InsertTraceListnerData(testingCtx, t, tt.tracelistenerData)

			utils.SetStoreData(testingCtx, tt.storeData)

			resp, err := http.Get(fmt.Sprintf(estimatePrimaryChannelsUrl, testingCtx.Cfg.ListenAddr))

			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			if !tt.success {
				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
				return
			}

			require.Equal(t, tt.expectedHttpCode, resp.StatusCode)

			if tt.expectedHttpCode == 200 {
				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				respStruct := chains.ChainsPrimaryChannelResponse{}
				err = json.Unmarshal(body, &respStruct)
				require.NoError(t, err)
				fmt.Println(respStruct)

				require.Equal(t, tt.expectedResponse.Chains["cosmos-hub"]["osmosis"], respStruct.Chains["cosmos-hub"]["osmosis"])
				require.Equal(t, tt.expectedResponse.Chains["osmosis"]["cosmos-hub"], respStruct.Chains["osmosis"]["cosmos-hub"])
			}
		})
	}
	utils.TruncateCNSDB(testingCtx, t)
}
