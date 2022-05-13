package database_test

import (
	"context"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
)

func (s *TestSuite) TestDenomTrace() {
	tests := []struct {
		name      string
		chainName string
		hash      string
		expRes    utils.DenomTrace
		success   bool
	}{
		{
			"chain not found",
			"invalidChain",
			"",
			utils.DenomTrace{},
			false,
		},
		{
			"disabled chain",
			utils.DisabledChain.ChainName,
			"",
			utils.DenomTrace{},
			false,
		},
		{
			"inserted chain but with invalid denom trace hash",
			utils.ChainWithPublicEndpoints.ChainName,
			"invalidhash",
			utils.VerifyTraceData.Denoms[0],
			false,
		},
		{
			"inserted chain with valid denom trace hash",
			utils.ChainWithoutPublicEndpoints.ChainName,
			utils.VerifyTraceData.Denoms[0].Hash,
			utils.VerifyTraceData.Denoms[0],
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.DenomTrace(context.Background(), tt.chainName, tt.hash)
			if tt.success {
				s.Require().NoError(err)
				s.Require().NotEmpty(res)
				s.Require().Equal(tt.expRes.ChainName, res.TracelistenerDatabaseRow.ChainName)
				s.Require().Equal(tt.expRes.BaseDenom, res.BaseDenom)
				s.Require().Equal(tt.expRes.Path, res.Path)
				s.Require().Equal(tt.expRes.Hash, res.Hash)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
