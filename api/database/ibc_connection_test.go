package database_test

import (
	utils "github.com/emerishq/demeris-api-server/api/test_utils"
)

func (s *TestSuite) TestConnection() {
	tests := []struct {
		name         string
		chainName    string
		connectionId string
		expRes       utils.Connection
		success      bool
	}{
		{
			"chain not found",
			"invalidChain",
			"",
			utils.Connection{},
			false,
		},
		{
			"inserted chain but with invalid connectionId",
			utils.ChainWithPublicEndpoints.ChainName,
			"invalidconnectionId",
			utils.VerifyTraceData.Connections[0],
			false,
		},
		{
			"inserted chain with valid connectionId",
			utils.ChainWithoutPublicEndpoints.ChainName,
			utils.VerifyTraceData.Connections[0].ConnectionID,
			utils.VerifyTraceData.Connections[0],
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.Connection(tt.chainName, tt.connectionId)
			if tt.success {
				s.Require().NoError(err)
				s.Require().NotEmpty(res)
				s.Require().Equal(tt.expRes.ChainName, res.TracelistenerDatabaseRow.ChainName)
				s.Require().Equal(tt.expRes.ClientID, res.ClientID)
				s.Require().Equal(tt.expRes.ConnectionID, res.ConnectionID)
				s.Require().Equal(tt.expRes.CounterClientID, res.CounterClientID)
				s.Require().Equal(tt.expRes.CounterConnectionID, res.CounterConnectionID)
				s.Require().Equal(tt.expRes.State, res.State)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
