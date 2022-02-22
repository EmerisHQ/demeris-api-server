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
	chainEndpointUrl  = "http://%s/chain/%s"
	chainsEndpointUrl = "http://%s/chains"
	chainStatusUrl    = "http://%s/chain/%s/status"
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
		{
			"Get Chain - Disabled",
			disabledChain,
			disabledChain.ChainName,
			400,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
			t.Parallel()

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

func toSupportedChain(c cns.Chain) chains.SupportedChain {

	return chains.SupportedChain{
		ChainName:   c.ChainName,
		DisplayName: c.DisplayName,
		Logo:        c.Logo,
	}
}

func TestGetChainStatus(t *testing.T) {

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
			chainWithoutPublicEndpoints,
			chainWithoutPublicEndpoints.ChainName,
			200,
			chains.StatusResponse{Online: false},
			true,
		},
		{
			"Get Chain Status - Enabled",
			chainWithPublicEndpoints,
			chainWithPublicEndpoints.ChainName,
			200,
			chains.StatusResponse{Online: true},
			true,
		},
		{
			"Get Chain Status - Disabled",
			disabledChain,
			disabledChain.ChainName,
			400,
			chains.StatusResponse{Online: false},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	utils.TruncateDB(testingCtx, t)
}
