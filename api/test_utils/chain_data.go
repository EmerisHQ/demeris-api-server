package test_utils

import (
	"time"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/lib/pq"
)

var relayerBalance = int64(30000)

var ChainWithoutPublicEndpoints = cns.Chain{
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
	ValidBlockThresh: cns.Threshold(30 * time.Minute),
	DerivationPath:   "m/44'/60'/0'/1",
	SupportedWallets: pq.StringArray([]string{"keplr"}),
	BlockExplorer:    "http://explorer.com",
}

var ChainWithPublicEndpoints = cns.Chain{
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
	ValidBlockThresh: cns.Threshold(30 * time.Minute),
	DerivationPath:   "m/44'/60'/0'/1",
	SupportedWallets: pq.StringArray([]string{"keplr"}),
	BlockExplorer:    "http://explorer.com",
	PublicNodeEndpoints: cns.PublicNodeEndpoints{
		TendermintRPC: []string{"https://www.host.com:1234"},
		CosmosAPI:     []string{"https://host.foo.bar:2345"},
	},
}

var VerifyTraceData = TracelistenerData{
	Denoms: []DenomTrace{
		{
			Path:      "transfer/ch1",
			BaseDenom: "denom2",
			Hash:      "abc12345",
			ChainName: "chain1",
		},
	},

	Channels: []Channel{
		{
			ChannelID:        "ch1",
			CounterChannelID: "ch2",
			Port:             "transfer",
			State:            3,
			Hops:             []string{"conn1", "conn2"},
			ChainName:        "chain1",
		},
		{
			ChannelID:        "ch2",
			CounterChannelID: "ch1",
			Port:             "transfer",
			State:            3,
			Hops:             []string{"conn2", "conn1"},
			ChainName:        "chain2",
		},
	},

	Connections: []Connection{
		{
			ChainName:           "chain1",
			ConnectionID:        "conn1",
			ClientID:            "cl1",
			State:               "ready",
			CounterConnectionID: "conn2",
			CounterClientID:     "cl2",
		},
		{
			ChainName:           "chain2",
			ConnectionID:        "conn2",
			ClientID:            "cl2",
			State:               "ready",
			CounterConnectionID: "conn1",
			CounterClientID:     "cl1",
		},
	},

	Clients: []Client{
		{
			SourceChainName: "chain1",
			DestChainID:     "chain_2",
			ClientID:        "cl1",
			LatestHeight:    "99",
			TrustingPeriod:  "10",
		},
		{
			SourceChainName: "chain2",
			DestChainID:     "chain_1",
			ClientID:        "cl2",
			LatestHeight:    "99",
			TrustingPeriod:  "10",
		},
	},

	BlockTimes: []BlockTime{
		{
			ChainName: "chain2",
			Time:      time.Now().UTC(),
		},
	},
}

var VerifyTraceData3Chains = TracelistenerData{
	Denoms: []DenomTrace{
		{
			Path:      "transfer/channel-11/transfer/channel-184",
			BaseDenom: "uakt",
			Hash:      "abc12345",
			ChainName: "regen",
		},
	},
	Channels: []Channel{
		{
			ChannelID:        "channel-11",
			CounterChannelID: "channel-185",
			Port:             "port",
			State:            3,
			Hops:             []string{"conn1", "conn2"},
			ChainName:        "regen",
		},
		{
			ChannelID:        "channel-185",
			CounterChannelID: "channel-11",
			Port:             "port",
			State:            3,
			Hops:             []string{"conn2", "conn1"},
			ChainName:        "cosmoshub",
		},
		{
			ChannelID:        "channel-184",
			CounterChannelID: "channel-17",
			Port:             "port",
			State:            3,
			Hops:             []string{"conn3", "conn4"},
			ChainName:        "cosmoshub",
		},
		{
			ChannelID:        "channel-17",
			CounterChannelID: "channel-184",
			Port:             "port",
			State:            3,
			Hops:             []string{"conn4", "conn3"},
			ChainName:        "akash",
		},
	},
	Connections: []Connection{
		{
			ChainName:           "regen",
			ConnectionID:        "conn1",
			ClientID:            "cl1",
			State:               "ready",
			CounterConnectionID: "conn2",
			CounterClientID:     "cl2",
		},
		{
			ChainName:           "cosmoshub",
			ConnectionID:        "conn2",
			ClientID:            "cl2",
			State:               "ready",
			CounterConnectionID: "conn2",
			CounterClientID:     "cl1",
		},
		{
			ChainName:           "cosmoshub",
			ConnectionID:        "conn3",
			ClientID:            "cl2",
			State:               "ready",
			CounterConnectionID: "conn4",
			CounterClientID:     "cl3",
		},
		{
			ChainName:           "akash",
			ConnectionID:        "conn4",
			ClientID:            "cl3",
			State:               "ready",
			CounterConnectionID: "conn3",
			CounterClientID:     "cl2",
		},
	},
	Clients: []Client{
		{
			SourceChainName: "regen",
			DestChainID:     "cosmoshub-4",
			ClientID:        "cl1",
			LatestHeight:    "99",
			TrustingPeriod:  "10",
		},
		{
			SourceChainName: "cosmoshub",
			DestChainID:     "regen-1",
			ClientID:        "cl2",
			LatestHeight:    "99",
			TrustingPeriod:  "10",
		},
		{
			SourceChainName: "cosmoshub",
			DestChainID:     "akashnet-2",
			ClientID:        "cl2",
			LatestHeight:    "99",
			TrustingPeriod:  "10",
		},
		{
			SourceChainName: "akash",
			DestChainID:     "cosmoshub-4",
			ClientID:        "cl3",
			LatestHeight:    "99",
			TrustingPeriod:  "10",
		},
	},
	BlockTimes: []BlockTime{
		{
			ChainName: "akash",
			Time:      time.Now().UTC(),
		},
	},
}

var DisabledChain = cns.Chain{
	Enabled:        false,
	ChainName:      "chain3",
	Logo:           "http://logo.com",
	DisplayName:    "Chain 3",
	PrimaryChannel: map[string]string{"key": "value"},
	Denoms: []cns.Denom{
		{
			Name:        "denom3",
			DisplayName: "Denom 3",
			Logo:        "http://logo.com",
			Precision:   8,
			Verified:    true,
			Stakable:    true,
			Ticker:      "DENOM3",
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
		ChainID:  "chain_123",
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
