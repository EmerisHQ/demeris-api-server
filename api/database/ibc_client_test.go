package database_test

import (
	utils "github.com/emerishq/demeris-api-server/api/test_utils"
)

func (s *TestSuite) TestQueryIBCClientTrace() {
	tests := []struct {
		name      string
		chainName string
		channel   string
		expLen    int
		success   bool
	}{
		{
			"chain not found",
			"invalidChain",
			"",
			0,
			false,
		},
		{
			"inserted chain but with invalid channel",
			utils.ChainWithPublicEndpoints.ChainName,
			"invalidchannel",
			0,
			false,
		},
		{
			"inserted chain with valid channel",
			utils.VerifyTraceData.Channels[0].ChainName,
			utils.VerifyTraceData.Channels[0].ChannelID,
			len(utils.VerifyTraceData.Channels[0].Hops),
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.QueryIBCClientTrace(tt.chainName, tt.channel)
			if tt.success {
				s.Require().NoError(err)
				s.Require().NotEmpty(res)
				s.Require().Len(res, tt.expLen)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
