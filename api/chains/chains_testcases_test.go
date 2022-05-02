package chains_test

import (
	"time"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/lib/pq"
)

var getChainTestCases = []struct {
	name             string
	dataStruct       cns.Chain
	chainName        string
	expectedHttpCode int
	success          bool
}{
	{
		"Get Chain - Unknown chain",
		cns.Chain{}, // ignored
		"foo",
		400,
		false,
	},
	{
		"Get Chain - Without PublicEndpoint",
		utils.ChainWithoutPublicEndpoints,
		utils.ChainWithoutPublicEndpoints.ChainName,
		200,
		true,
	},
	{
		"Get Chain - With PublicEndpoints",
		utils.ChainWithPublicEndpoints,
		utils.ChainWithPublicEndpoints.ChainName,
		200,
		true,
	},
}

type testChainWithStatus struct {
	chain  cns.Chain
	online bool
}

var getChainsTestCases = []struct {
	name             string
	dataStruct       []testChainWithStatus
	expectedHttpCode int
	success          bool
}{
	{
		"Get Chains - Nothing in DB",
		[]testChainWithStatus{}, // ignored
		200,
		true,
	},
	{
		"Get Chains - 2 Chains (With & Without)",
		[]testChainWithStatus{
			{
				chain:  utils.ChainWithoutPublicEndpoints,
				online: false,
			},
			{
				chain:  utils.ChainWithPublicEndpoints,
				online: true,
			},
		},
		200,
		true,
	},
}

var verifyTraceTestCases = []struct {
	name             string
	dataStruct       utils.TracelistenerData
	chains           []cns.Chain
	sourceChain      string
	hash             string
	cause            string
	verified         bool
	expectedHttpCode int
}{
	{
		"chain1->ch1->Chain2",
		utils.VerifyTraceData,
		[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		"chain1",
		"abc12345",
		"",
		true,
		200,
	},
	{
		"wrong hash",
		utils.VerifyTraceData,
		[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		"chain1",
		"xyz",
		"token hash xyz not found on chain chain1",
		false,
		200,
	},
	{
		"denom doesn't exist on dest chain",
		utils.TracelistenerData{
			Denoms: []utils.DenomTrace{
				{
					Path:      "transfer/ch1",
					BaseDenom: "denomXYZ",
					Hash:      "12345",
					ChainName: "chain1",
				},
			},
			Channels:    utils.VerifyTraceData.Channels,
			Connections: utils.VerifyTraceData.Connections,
			Clients:     utils.VerifyTraceData.Clients,
			BlockTimes:  utils.VerifyTraceData.BlockTimes,
		},
		[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		"chain1",
		"abc12345",
		"",
		false,
		200,
	},
	{
		"incorrect channel name in path",
		utils.TracelistenerData{
			Denoms: []utils.DenomTrace{
				{
					Path:      "transfer/ch2",
					BaseDenom: "denom2",
					Hash:      "abc12345",
					ChainName: "chain1",
				},
			},
			Channels:    utils.VerifyTraceData.Channels,
			Connections: utils.VerifyTraceData.Connections,
			Clients:     utils.VerifyTraceData.Clients,
			BlockTimes:  utils.VerifyTraceData.BlockTimes,
		},
		[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		"chain1",
		"abc12345",
		"no destination chain found",
		false,
		200,
	},
	{
		"no matching connection id",
		utils.TracelistenerData{
			Denoms:   utils.VerifyTraceData.Denoms,
			Channels: utils.VerifyTraceData.Channels,
			Connections: []utils.Connection{
				{
					ChainName:           "chain1",
					ConnectionID:        "testconn",
					ClientID:            "cl1",
					State:               "ready",
					CounterConnectionID: "conn2",
					CounterClientID:     "cl2",
				},
				utils.VerifyTraceData.Connections[1],
			},
			Clients:    utils.VerifyTraceData.Clients,
			BlockTimes: utils.VerifyTraceData.BlockTimes,
		},
		[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		"chain1",
		"abc12345",
		"no destination chain found",
		false,
		200,
	},
	{
		"Channels.hops incorrect conn",
		utils.TracelistenerData{
			Denoms: utils.VerifyTraceData.Denoms,
			Channels: []utils.Channel{
				{
					ChannelID:        "ch1",
					CounterChannelID: "ch2",
					Hops:             []string{},
					ChainName:        "chain1",
				},
				{
					ChannelID:        "ch2",
					CounterChannelID: "ch1",
					Hops:             []string{},
					ChainName:        "chain2",
				},
			},
			Connections: utils.VerifyTraceData.Connections,
			Clients:     utils.VerifyTraceData.Clients,
			BlockTimes:  utils.VerifyTraceData.BlockTimes,
		},
		[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		"chain1",
		"abc12345",
		"no destination chain found",
		false,
		200,
	},
	{
		"destination chain doesn't exist",
		utils.VerifyTraceData,
		[]cns.Chain{utils.ChainWithoutPublicEndpoints},
		"chain1",
		"abc12345",
		"no chain with name chain2 found",
		false,
		200,
	},
	{
		"dest chain not enabled",
		utils.VerifyTraceData,
		[]cns.Chain{
			utils.ChainWithoutPublicEndpoints,
			{
				Enabled:          false,
				ChainName:        "chain2",
				DemerisAddresses: []string{"12345"},
				SupportedWallets: pq.StringArray([]string{"keplr"}),
				PrimaryChannel:   map[string]string{"chain2": "chXYZ"},
				Denoms: []cns.Denom{
					{
						Name:     "denom2",
						Verified: true,
					},
				},
				NodeInfo: cns.NodeInfo{
					ChainID: "chain_2",
				},
			},
		},
		"chain1",
		"abc12345",
		"no chain with name chain2 found",
		false,
		200,
	},
	{
		"dest chain offline",
		utils.TracelistenerData{
			Denoms:      utils.VerifyTraceData.Denoms,
			Channels:    utils.VerifyTraceData.Channels,
			Connections: utils.VerifyTraceData.Connections,
			Clients:     utils.VerifyTraceData.Clients,
			BlockTimes: []utils.BlockTime{
				{
					ChainName: "chain2",
					Time:      time.Now().Add(time.Hour * -24),
				},
			},
		},
		[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		"chain1",
		"abc12345",
		"status offline",
		false,
		200,
	},
	{
		"primary channel mismatch in chains data",
		utils.VerifyTraceData,
		[]cns.Chain{
			utils.ChainWithPublicEndpoints,
			{
				Enabled:          true,
				ChainName:        "chain1",
				DemerisAddresses: []string{"12345"},
				SupportedWallets: pq.StringArray([]string{"keplr"}),
				PrimaryChannel:   map[string]string{"chain2": "chXYZ"},
				Denoms: []cns.Denom{
					{
						Name:     "denom1",
						Verified: true,
					},
				},
				NodeInfo: cns.NodeInfo{
					ChainID: "chain_1",
				},
			},
		},
		"chain1",
		"abc12345",
		"",
		true,
		200,
	},
	{
		"primary channel mismatch in chains data",
		utils.VerifyTraceData,
		[]cns.Chain{
			utils.ChainWithPublicEndpoints,
			{
				Enabled:          true,
				ChainName:        "chain1",
				DemerisAddresses: []string{"12345"},
				SupportedWallets: pq.StringArray([]string{"keplr"}),
				PrimaryChannel:   map[string]string{"chain1": "ch1"},
				Denoms: []cns.Denom{
					{
						Name:     "denom1",
						Verified: true,
					},
				},
				NodeInfo: cns.NodeInfo{
					ChainID: "chain_1",
				},
			},
		},
		"chain1",
		"abc12345",
		"",
		true,
		200,
	},
	{
		"akash->cosmoshub->regen",
		utils.VerifyTraceData3Chains,
		[]cns.Chain{
			{
				Enabled:          true,
				ChainName:        "akash",
				DemerisAddresses: []string{"12345"},
				SupportedWallets: pq.StringArray([]string{"keplr"}),
				ValidBlockThresh: cns.Threshold(30 * time.Minute),
				Denoms: []cns.Denom{
					{
						Name:     "uakt",
						Verified: true,
					},
				},
				NodeInfo: cns.NodeInfo{
					ChainID: "akashnet-2",
				},
			},
			{
				Enabled:          true,
				ChainName:        "cosmoshub",
				DemerisAddresses: []string{"12345"},
				SupportedWallets: pq.StringArray([]string{"keplr"}),
				ValidBlockThresh: cns.Threshold(30 * time.Minute),
				Denoms: []cns.Denom{
					{
						Name:     "uatom",
						Verified: true,
					},
				},
				NodeInfo: cns.NodeInfo{
					ChainID: "cosmoshub-4",
				},
			},
			{
				Enabled:          true,
				ChainName:        "regen",
				DemerisAddresses: []string{"12345"},
				SupportedWallets: pq.StringArray([]string{"keplr"}),
				ValidBlockThresh: cns.Threshold(30 * time.Minute),
				Denoms: []cns.Denom{
					{
						Name:     "uregen",
						Verified: true,
					},
				},
				NodeInfo: cns.NodeInfo{
					ChainID: "regen-1",
				},
			},
		},
		"regen",
		"abc12345",
		"",
		true,
		200,
	},
}
