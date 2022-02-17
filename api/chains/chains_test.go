package chains_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/allinbits/demeris-api-server/api/chains"
	utils "github.com/allinbits/demeris-api-server/api/test_utils"

	"github.com/allinbits/demeris-backend-models/cns"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

const (
	chainEndpointUrl       = "http://%s/chain/%s"
	chainsEndpointUrl      = "http://%s/chains"
	verifyTraceEndpointUrl = "http://%s/chain/%s/denom/verify_trace/%s"
)

func TestGetChain(t *testing.T) {

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// arrange
			// if we have a populated Chain store, add it
			if !cmp.Equal(tt.dataStruct, cns.Chain{}) {
				err := testingCtx.CnsDB.AddChain(tt.dataStruct)
				require.NoError(t, err)
			}

			// act
			resp, err := http.Get(fmt.Sprintf(chainEndpointUrl, testingCtx.Cfg.ListenAddr, tt.chainName))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// assert
			if !tt.success {
				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			respStruct := chains.ChainResponse{}
			err = json.Unmarshal(body, &respStruct)
			require.NoError(t, err)

			require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
			require.Equal(t, tt.dataStruct, respStruct.Chain)
		})
	}
	utils.TruncateDB(testingCtx, t)
}

func TestGetChains(t *testing.T) {

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// arrange
			// if we have a populated Chain store, add it
			if len(tt.dataStruct) != 0 {
				for _, c := range tt.dataStruct {
					err := testingCtx.CnsDB.AddChain(c)
					require.NoError(t, err)
				}
			}

			// act
			resp, err := http.Get(fmt.Sprintf(chainsEndpointUrl, testingCtx.Cfg.ListenAddr))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// assert
			if !tt.success {
				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				respStruct := chains.ChainsResponse{}
				err = json.Unmarshal(body, &respStruct)
				require.NoError(t, err)

				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
				for _, c := range tt.dataStruct {
					require.Contains(t, respStruct.Chains, toSupportedChain(c))
				}
			}
		})
	}
	utils.TruncateDB(testingCtx, t)
}

func TestVerifyTrace(t *testing.T) {

	tests := []struct {
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
		// {
		// 	"Channels.hops incorrect conn",
		// 	tracelistenerData{
		// 		denoms: verifyTraceData.denoms,
		// 		channels: []channel{
		// 			{
		// 				channelID:        "ch1",
		// 				counterChannelID: "ch2",
		// 				hops:             []string{},
		// 				chainName:        "chain1",
		// 			},
		// 			{
		// 				channelID:        "ch2",
		// 				counterChannelID: "ch1",
		// 				hops:             []string{},
		// 				chainName:        "chain2",
		// 			},
		// 		},
		// 		connections: verifyTraceData.connections,
		// 		clients:     verifyTraceData.clients,
		// 		blockTimes:  verifyTraceData.blockTimes,
		// 	},
		// 	[]cns.Chain{chainWithPublicEndpoints, chainWithoutPublicEndpoints},
		// 	"chain1",
		// 	"12345",
		// 	"",
		// 	false,
		// 	200,
		// },
	}

	runTraceListnerMigrations(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insertTraceListnerData(t, tt.dataStruct)
			for _, chain := range tt.chains {
				testingCtx.CnsDB.AddChain(chain)

			}

			resp, err := http.Get(fmt.Sprintf(verifyTraceEndpointUrl, testingCtx.Cfg.ListenAddr, tt.sourceChain, tt.hash))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tt.expectedHttpCode, resp.StatusCode)

			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			var data map[string]map[string]interface{}
			err = json.Unmarshal(b, &data)
			require.NoError(t, err)

			result := data["verify_trace"]
			// fmt.Println(result)

			if tt.cause != "" {
				require.Contains(t, result["cause"], tt.cause)
			}

			require.Equal(t, tt.verified, result["verified"])
		})
	}
}

func toSupportedChain(c cns.Chain) chains.SupportedChain {

	return chains.SupportedChain{
		ChainName:   c.ChainName,
		DisplayName: c.DisplayName,
		Logo:        c.Logo,
	}
}
