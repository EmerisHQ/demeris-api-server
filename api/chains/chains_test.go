package chains_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/emerishq/demeris-api-server/api/chains"
	utils "github.com/emerishq/demeris-api-server/api/test_utils"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

const (
	chainEndpointUrl       = "http://%s/chain/%s"
	chainsEndpointUrl      = "http://%s/chains"
	chainsStatusesUrl      = "http://%s/chains/status"
	chainStatusUrl         = "http://%s/chain/%s/status"
	chainSupplyUrl         = "http://%s/chain/%s/supply"
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
		{
			"Get Chain - Disabled",
			utils.DisabledChain,
			utils.DisabledChain.ChainName,
			400,
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

			require.Equal(t, tt.expectedHttpCode, resp.StatusCode)

			if tt.expectedHttpCode == 200 {
				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				respStruct := chains.ChainResponse{}
				err = json.Unmarshal(body, &respStruct)
				require.NoError(t, err)

				require.Equal(t, tt.dataStruct, respStruct.Chain)
			}
		})
	}
	utils.TruncateCNSDB(testingCtx, t)
}

func TestGetChains(t *testing.T) {
	utils.RunTraceListnerMigrations(testingCtx, t)
	utils.InsertTraceListnerData(testingCtx, t, utils.VerifyTraceData)

	for _, tt := range getChainsTestCases {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			// if we have a populated Chain store, add it
			if len(tt.dataStruct) != 0 {
				for _, c := range tt.dataStruct {
					err := testingCtx.CnsDB.AddChain(c.chain)
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
					require.Contains(t, respStruct.Chains, utils.ToChainWithStatus(c.chain, c.online))
				}
			}
		})
	}
	utils.TruncateCNSDB(testingCtx, t)
}

func TestVerifyTrace(t *testing.T) {
	utils.RunTraceListnerMigrations(testingCtx, t)

	for i, tt := range verifyTraceTestCases {
		t.Run(fmt.Sprintf("%d %s", i, tt.name), func(t *testing.T) {
			utils.InsertTraceListnerData(testingCtx, t, tt.dataStruct)
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

			if result["hash"] != nil {
				h := result["hash"].(string)
				require.Equal(t, "ibc/", h[:4])
				require.Equal(t, strings.ToUpper(tt.hash), h[4:])
			}

			require.Equal(t, tt.verified, result["verified"], "result cause=%s", result["cause"])
		})
		utils.TruncateTracelistener(testingCtx, t)
		utils.TruncateCNSDB(testingCtx, t)
	}
}

func TestGetChainStatus(t *testing.T) {
	utils.RunTraceListnerMigrations(testingCtx, t)
	utils.InsertTraceListnerData(testingCtx, t, utils.VerifyTraceData)

	tests := []struct {
		name             string
		dataStruct       cns.Chain
		chainName        string
		expectedHttpCode int
		expectedResponse chains.StatusResponse
		success          bool
	}{
		{
			"Get Chain Status - Without PublicEndpoint",
			utils.ChainWithoutPublicEndpoints,
			utils.ChainWithoutPublicEndpoints.ChainName,
			200,
			chains.StatusResponse{Online: false},
			true,
		},
		{
			"Get Chain Status - Enabled",
			utils.ChainWithPublicEndpoints,
			utils.ChainWithPublicEndpoints.ChainName,
			200,
			chains.StatusResponse{Online: true},
			true,
		},
		{
			"Get Chain Status - Disabled",
			utils.DisabledChain,
			utils.DisabledChain.ChainName,
			400,
			chains.StatusResponse{Online: false},
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
			resp, err := http.Get(fmt.Sprintf(chainStatusUrl, testingCtx.Cfg.ListenAddr, tt.chainName))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// assert
			if !tt.success {
				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			respStruct := chains.StatusResponse{}
			err = json.Unmarshal(body, &respStruct)
			require.NoError(t, err)

			require.Equal(t, tt.expectedResponse, respStruct)

			require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
		})
	}
	utils.TruncateCNSDB(testingCtx, t)
}

func TestGetChainSupply(t *testing.T) {
	tests := []struct {
		name             string
		dataStruct       cns.Chain
		chainName        string
		expectedHttpCode int
		expectedResponse chains.SupplyResponse
		success          bool
	}{
		{
			"Get Chain Supply - Enabled",
			utils.ChainWithPublicEndpoints,
			utils.ChainWithPublicEndpoints.ChainName,
			500,
			chains.SupplyResponse{Supply: []chains.Coin(nil), Pagination: chains.Pagination{}},
			false,
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
			resp, err := http.Get(fmt.Sprintf(chainSupplyUrl, testingCtx.Cfg.ListenAddr, tt.chainName))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// assert
			if !tt.success {
				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			respStruct := chains.SupplyResponse{}
			err = json.Unmarshal(body, &respStruct)
			require.NoError(t, err)

			require.Equal(t, tt.expectedResponse, respStruct)

			require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
		})
	}
	utils.TruncateCNSDB(testingCtx, t)
}

func TestGetChainsStatuses(t *testing.T) {
	utils.RunTraceListnerMigrations(testingCtx, t)
	utils.InsertTraceListnerData(testingCtx, t, utils.VerifyTraceData)

	// arrange
	testChains := []cns.Chain{
		utils.ChainWithoutPublicEndpoints,
		utils.ChainWithPublicEndpoints,
		utils.DisabledChain,
	}
	for _, c := range testChains {
		err := testingCtx.CnsDB.AddChain(c)
		require.NoError(t, err)
	}

	// act
	resp, err := http.Get(fmt.Sprintf(chainsStatusesUrl, testingCtx.Cfg.ListenAddr))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// assert
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	respStruct := chains.ChainsStatusesResponse{}
	err = json.Unmarshal(body, &respStruct)
	require.NoError(t, err)

	expectedResult := chains.ChainsStatusesResponse{
		Chains: map[string]chains.ChainStatus{
			utils.ChainWithoutPublicEndpoints.ChainName: {
				Online: false,
			},
			utils.ChainWithPublicEndpoints.ChainName: {
				Online: true,
			},
		},
	}
	require.Equal(t, expectedResult, respStruct)
	require.Equal(t, 200, resp.StatusCode)

	utils.TruncateCNSDB(testingCtx, t)
}
