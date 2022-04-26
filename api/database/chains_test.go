package database_test

import (
	"encoding/json"
	"testing"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestChain() {
	tests := []struct {
		name      string
		chainName string
		expRes    cns.Chain
		success   bool
	}{
		{
			"chain not found",
			"invalidChain",
			cns.Chain{},
			false,
		},
		{
			"disabled chain",
			utils.DisabledChain.ChainName,
			cns.Chain{},
			false,
		},
		{
			"inserted chain",
			utils.ChainWithPublicEndpoints.ChainName,
			utils.ChainWithPublicEndpoints,
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.Chain(tt.chainName)
			if tt.success {
				s.Require().NoError(err)
				compareAsBytes(s.T(), tt.expRes, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func compareAsBytes(t *testing.T, expected interface{}, actual interface{}) {
	actualBytes, err := json.Marshal(actual)
	require.NoError(t, err)
	expBytes, err := json.Marshal(expected)
	require.NoError(t, err)
	require.Equal(t, expBytes, actualBytes)
}

func (s *TestSuite) TestChainExists() {
	tests := []struct {
		name      string
		chainName string
		expRes    bool
	}{
		{
			"chain not found",
			"invalidChain",
			false,
		},
		{
			"disabled chain",
			utils.DisabledChain.ChainName,
			false,
		},
		{
			"inserted chain",
			utils.ChainWithPublicEndpoints.ChainName,
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.ChainExists(tt.chainName)
			s.Require().NoError(err)
			s.Require().Equal(tt.expRes, res)
		})
	}
}

func (s *TestSuite) TestChainFromChainID() {
	tests := []struct {
		name    string
		chainID string
		expRes  cns.Chain
		success bool
	}{
		{
			"chain not found",
			"invalidChain",
			cns.Chain{},
			false,
		},
		{
			"disabled chain",
			utils.DisabledChain.NodeInfo.ChainID,
			cns.Chain{},
			false,
		},
		{
			"inserted chain",
			utils.ChainWithPublicEndpoints.NodeInfo.ChainID,
			utils.ChainWithPublicEndpoints,
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.ChainFromChainID(tt.chainID)
			if tt.success {
				s.Require().NoError(err)
				compareAsBytes(s.T(), tt.expRes, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
