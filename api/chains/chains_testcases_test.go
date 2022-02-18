package chains_test

import (
	"time"

	"github.com/allinbits/demeris-backend-models/cns"
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
		chainWithoutPublicEndpoints,
		chainWithoutPublicEndpoints.ChainName,
		200,
		true,
	},
	{
		"Get Chain - With PublicEndpoints",
		chainWithPublicEndpoints,
		chainWithPublicEndpoints.ChainName,
		200,
		true,
	},
}

var getChainsTestCases = []struct {
	name             string
	dataStruct       []cns.Chain
	expectedHttpCode int
	success          bool
}{
	{
		"Get Chains - Nothing in DB",
		[]cns.Chain{}, // ignored
		200,
		true,
	},
	{
		"Get Chains - 2 Chains (With & Without)",
		[]cns.Chain{chainWithoutPublicEndpoints, chainWithPublicEndpoints},
		200,
		true,
	},
}

var verifyTraceTestCases = []struct {
	name             string
	dataStruct       tracelistenerData
	chains           []cns.Chain
	sourceChain      string
	hash             string
	cause            string
	verified         bool
	expectedHttpCode int
}{
	{
		"chain1->ch1->Chain2",
		verifyTraceData,
		[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		"chain1",
		"12345",
		"",
		true,
		200,
	},
	{
		"wrong hash",
		verifyTraceData,
		[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		"chain1",
		"xyz",
		"token hash xyz not found on chain chain1",
		false,
		200,
	},
	{
		"denom doesn't exist on dest chain",
		tracelistenerData{
			denoms: []denomTrace{
				{
					path:      "transfer/ch1",
					baseDenom: "denomXYZ",
					hash:      "12345",
					chainName: "chain1",
				},
			},
			channels:    verifyTraceData.channels,
			connections: verifyTraceData.connections,
			clients:     verifyTraceData.clients,
			blockTimes:  verifyTraceData.blockTimes,
		},
		[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		"chain1",
		"12345",
		"",
		false,
		200,
	},
	{
		"incorrect channel name in path",
		tracelistenerData{
			denoms: []denomTrace{
				{
					path:      "transfer/ch2",
					baseDenom: "denom2",
					hash:      "12345",
					chainName: "chain1",
				},
			},
			channels:    verifyTraceData.channels,
			connections: verifyTraceData.connections,
			clients:     verifyTraceData.clients,
			blockTimes:  verifyTraceData.blockTimes,
		},
		[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		"chain1",
		"12345",
		"no destination chain found",
		false,
		200,
	},
	{
		"no matching connection id",
		tracelistenerData{
			denoms:   verifyTraceData.denoms,
			channels: verifyTraceData.channels,
			connections: []connection{
				{
					chainName:           "chain1",
					connectionID:        "testconn",
					clientID:            "cl1",
					state:               "ready",
					counterConnectionID: "conn2",
					counterClientID:     "cl2",
				},
				verifyTraceData.connections[1],
			},
			clients:    verifyTraceData.clients,
			blockTimes: verifyTraceData.blockTimes,
		},
		[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		"chain1",
		"12345",
		"no destination chain found",
		false,
		200,
	},
	{
		"Channels.hops incorrect conn",
		tracelistenerData{
			denoms: verifyTraceData.denoms,
			channels: []channel{
				{
					channelID:        "ch1",
					counterChannelID: "ch2",
					hops:             []string{},
					chainName:        "chain1",
				},
				{
					channelID:        "ch2",
					counterChannelID: "ch1",
					hops:             []string{},
					chainName:        "chain2",
				},
			},
			connections: verifyTraceData.connections,
			clients:     verifyTraceData.clients,
			blockTimes:  verifyTraceData.blockTimes,
		},
		[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		"chain1",
		"12345",
		"no destination chain found",
		false,
		200,
	},
	{
		"destination chain doesn't exist",
		verifyTraceData,
		[]cns.Chain{chainWithoutPublicEndpoints},
		"chain1",
		"12345",
		"no chain with name chain2 found",
		false,
		200,
	},
	{
		"dest chain not enabled",
		verifyTraceData,
		[]cns.Chain{
			chainWithoutPublicEndpoints,
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
		"12345",
		"no chain with name chain2 found",
		false,
		200,
	},
	{
		"dest chain offline",
		tracelistenerData{
			denoms:      verifyTraceData.denoms,
			channels:    verifyTraceData.channels,
			connections: verifyTraceData.connections,
			clients:     verifyTraceData.clients,
			blockTimes: []blockTime{
				{
					chainName: "chain2",
					time:      time.Now().Add(time.Hour * -24),
				},
			},
		},
		[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		"chain1",
		"12345",
		"status offline",
		false,
		200,
	},
	{
		"primary channel mismatch in chains data",
		verifyTraceData,
		[]cns.Chain{
			chainWithPublicEndpoints,
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
		"12345",
		"not primary channel for chain",
		false,
		200,
	},
	{
		"primary channel mismatch in chains data",
		verifyTraceData,
		[]cns.Chain{
			chainWithPublicEndpoints,
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
		"12345",
		"chain1 doesnt have primary channel for chain2",
		false,
		200,
	},
}
