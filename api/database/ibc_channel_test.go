package database_test

import (
	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/emerishq/demeris-backend-models/cns"
)

func (s *TestSuite) TestGetIbcChannelToChain() {
	tests := []struct {
		name      string
		chainName string
		channel   string
		chainID   string
		expRes    cns.IbcChannelsInfo
		success   bool
	}{
		{
			"chain not found",
			"invalidChain",
			"",
			"",
			cns.IbcChannelsInfo{},
			false,
		},
		{
			"inserted chain but with invalid channel",
			utils.ChainWithPublicEndpoints.ChainName,
			"invalidchannel",
			"",
			cns.IbcChannelsInfo{},
			false,
		},
		{
			"inserted chain with valid channel",
			utils.ChainWithoutPublicEndpoints.ChainName,
			utils.VerifyTraceData.Channels[0].ChannelID,
			utils.ChainWithoutPublicEndpoints.NodeInfo.ChainID,
			cns.IbcChannelsInfo{
				cns.IbcChannelInfo{
					ChainAName:             utils.VerifyTraceData.Channels[0].ChainName,
					ChainAChannelID:        utils.VerifyTraceData.Channels[0].ChannelID,
					ChainACounterChannelID: utils.VerifyTraceData.Channels[0].CounterChannelID,
					ChainAChainID:          utils.VerifyTraceData.Clients[0].DestChainID,
					ChainBName:             utils.VerifyTraceData.Channels[1].ChainName,
					ChainBChannelID:        utils.VerifyTraceData.Channels[1].ChannelID,
					ChainBCounterChannelID: utils.VerifyTraceData.Channels[1].CounterChannelID,
					ChainBChainID:          utils.VerifyTraceData.Clients[1].DestChainID,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.GetIbcChannelToChain(tt.chainName, tt.channel, tt.chainID)
			if tt.success {
				s.Require().NoError(err)
				s.Require().NotEmpty(res)
				s.Require().Equal(tt.expRes, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
