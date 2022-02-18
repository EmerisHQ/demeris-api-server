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
	for _, tt := range getChainTestCases {
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
	for _, tt := range getChainsTestCases {
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
	t.Parallel()

	runTraceListnerMigrations(t)

	for i, tt := range verifyTraceTestCases {
		t.Run(fmt.Sprintf("%d %s", i, tt.name), func(t *testing.T) {
			insertTraceListnerData(t, tt.dataStruct)
			for _, chain := range tt.chains {
				require.NoError(t, testingCtx.CnsDB.AddChain(chain))
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

			if tt.cause != "" {
				require.Contains(t, result["cause"], tt.cause)
			}

			require.Equal(t, tt.verified, result["verified"])
		})
		truncateTracelistener(t)
		utils.TruncateDB(testingCtx, t)
	}
}

func toSupportedChain(c cns.Chain) chains.SupportedChain {

	return chains.SupportedChain{
		ChainName:   c.ChainName,
		DisplayName: c.DisplayName,
		Logo:        c.Logo,
	}
}
