package chains_test

import (
	"time"

	"github.com/allinbits/demeris-backend-models/cns"
	"github.com/lib/pq"
)

type tracelistenerData struct {
	denoms      []denomTrace
	channels    []channel
	connections []connection
	clients     []client
	blockTimes  []blockTime
}

type denomTrace struct {
	path      string
	baseDenom string
	hash      string
	chainName string
}

type channel struct {
	channelID        string
	counterChannelID string
	port             string
	state            int
	hops             []string
	chainName        string
}

type connection struct {
	chainName           string
	connectionID        string
	clientID            string
	state               string
	counterConnectionID string
	counterClientID     string
}

type client struct {
	sourceChainName string
	destChainID     string
	clientID        string
	latestHeight    string
	trustingPeriod  string
}

type blockTime struct {
	chainName string
	time      time.Time
}

var relayerBalance = int64(30000)

var chainWithoutPublicEndpoints = cns.Chain{
	Enabled:        true,
	ChainName:      "chain1",
	Logo:           "http://logo.com",
	DisplayName:    "Chain 1",
	PrimaryChannel: map[string]string{"key": "value", "chain2": "ch1"},
	Denoms: []cns.Denom{
		{
			Name:        "denom1",
			DisplayName: "Denom 1",
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
		ChainID:  "chain_1",
		Bech32Config: cns.Bech32Config{
			MainPrefix:      "prefix",
			PrefixAccount:   "acc",
			PrefixValidator: "val",
			PrefixConsensus: "cons",
			PrefixPublic:    "pub",
			PrefixOperator:  "oper",
		},
	},
	ValidBlockThresh: cns.Threshold(30),
	DerivationPath:   "m/44'/60'/0'/1",
	SupportedWallets: pq.StringArray([]string{"keplr"}),
	BlockExplorer:    "http://explorer.com",
}

var chainWithPublicEndpoints = cns.Chain{
	Enabled:        true,
	ChainName:      "chain2",
	Logo:           "http://logo.com",
	DisplayName:    "Chain 2",
	PrimaryChannel: map[string]string{"key": "value"},
	Denoms: []cns.Denom{
		{
			Name:        "denom2",
			DisplayName: "Denom 2",
			Logo:        "http://logo.com",
			Precision:   8,
			Verified:    true,
			Stakable:    true,
			Ticker:      "DENOM2",
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
		ChainID:  "chain_2",
		Bech32Config: cns.Bech32Config{
			MainPrefix:      "prefix",
			PrefixAccount:   "acc",
			PrefixValidator: "val",
			PrefixConsensus: "cons",
			PrefixPublic:    "pub",
			PrefixOperator:  "oper",
		},
	},
	ValidBlockThresh: cns.Threshold(30),
	DerivationPath:   "m/44'/60'/0'/1",
	SupportedWallets: pq.StringArray([]string{"keplr"}),
	BlockExplorer:    "http://explorer.com",
	PublicNodeEndpoints: cns.PublicNodeEndpoints{
		TendermintRPC: []string{"https://www.host.com:1234"},
		CosmosAPI:     []string{"https://host.foo.bar:2345"},
	},
}

var verifyTraceData = tracelistenerData{
	denoms: []denomTrace{
		{
			path:      "transfer/ch1",
			baseDenom: "denom2",
			hash:      "12345",
			chainName: "chain1",
		},
	},

	channels: []channel{
		{
			channelID:        "ch1",
			counterChannelID: "ch2",
			port:             "transfer",
			state:            3,
			hops:             []string{"conn1", "conn2"},
			chainName:        "chain1",
		},
		{
			channelID:        "ch2",
			counterChannelID: "ch1",
			port:             "transfer",
			state:            3,
			hops:             []string{"conn2", "conn1"},
			chainName:        "chain2",
		},
	},

	connections: []connection{
		{
			chainName:           "chain1",
			connectionID:        "conn1",
			clientID:            "cl1",
			state:               "ready",
			counterConnectionID: "conn2",
			counterClientID:     "cl2",
		},
		{
			chainName:           "chain2",
			connectionID:        "conn2",
			clientID:            "cl2",
			state:               "ready",
			counterConnectionID: "conn1",
			counterClientID:     "cl1",
		},
	},

	clients: []client{
		{
			sourceChainName: "chain1",
			destChainID:     "chain_2",
			clientID:        "cl1",
			latestHeight:    "99",
			trustingPeriod:  "10",
		},
		{
			sourceChainName: "chain2",
			destChainID:     "chain_1",
			clientID:        "cl2",
			latestHeight:    "99",
			trustingPeriod:  "10",
		},
	},

	blockTimes: []blockTime{
		{
			chainName: "chain2",
			time:      time.Now(),
		},
	},
}
