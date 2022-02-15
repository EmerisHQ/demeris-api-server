package chains_test

import (
	"github.com/allinbits/demeris-backend-models/cns"
	"github.com/lib/pq"
)

var relayerBalance = int64(30000)

var chainWithoutPublicEndpoints = cns.Chain{
	Enabled:        true,
	ChainName:      "chain1",
	Logo:           "http://logo.com",
	DisplayName:    "Chain 1",
	PrimaryChannel: map[string]string{"key": "value"},
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

var disabledChain = cns.Chain{
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
