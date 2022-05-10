package database_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/emerishq/demeris-api-server/api/database"
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

// marshalling to json as we cannot compare ID value getting from DB which is dynamic
// and it will be ignored when converted to bytes.
func compareAsBytes(t *testing.T, expected interface{}, actual interface{}) {
	actualBytes, err := json.Marshal(actual)
	require.NoError(t, err)
	expBytes, err := json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expBytes), string(actualBytes))
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

func (s *TestSuite) TestChainLastBlock() {
	tests := []struct {
		name      string
		chainName string
		expRes    time.Time
	}{
		{
			"chain not found",
			"invalidChain",
			time.Time{},
		},
		{
			"disabled chain",
			utils.DisabledChain.ChainName,
			time.Time{},
		},
		{
			"inserted chain but tracelistener blocktime data not found",
			utils.ChainWithoutPublicEndpoints.ChainName,
			time.Time{},
		},
		{
			"inserted chain and racelistener blocktime data found",
			utils.ChainWithPublicEndpoints.ChainName,
			utils.VerifyTraceData.BlockTimes[0].Time,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.ChainLastBlock(tt.chainName)
			s.Require().NoError(err)
			s.Require().Equal(tt.expRes.Unix(), res.BlockTime.Unix())
		})
	}
}

func (s *TestSuite) TestChains() {
	tests := []struct {
		name   string
		expRes []cns.Chain
	}{
		{
			"enabled chains",
			[]cns.Chain{utils.ChainWithPublicEndpoints, utils.ChainWithoutPublicEndpoints},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.Chains()
			s.Require().NoError(err)
			s.Require().NotEmpty(res)
			compareAsBytes(s.T(), tt.expRes, res)
		})
	}
}

func (s *TestSuite) TestVerifiedDenoms() {
	tests := []struct {
		name   string
		expRes map[string]cns.DenomList
	}{
		{
			"enabled chains - verified denoms",
			map[string]cns.DenomList{
				utils.ChainWithPublicEndpoints.ChainName:    utils.ChainWithPublicEndpoints.VerifiedTokens(),
				utils.ChainWithoutPublicEndpoints.ChainName: utils.ChainWithoutPublicEndpoints.VerifiedTokens(),
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.VerifiedDenoms()
			s.Require().NoError(err)
			s.Require().NotEmpty(res)
			s.Require().Equal(tt.expRes, res)
		})
	}
}

func (s *TestSuite) TestChainsWithStatus() {
	tests := []struct {
		name   string
		expRes []database.ChainWithStatus
	}{
		{
			"enabled chains",
			[]database.ChainWithStatus{
				utils.ToChainWithStatus(utils.ChainWithPublicEndpoints, true),
				utils.ToChainWithStatus(utils.ChainWithoutPublicEndpoints, false),
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.ChainsWithStatus()
			s.Require().NoError(err)
			s.Require().NotEmpty(res)
			s.Require().Equal(tt.expRes, res)
		})
	}
}

func (s *TestSuite) TestChainIDs() {
	tests := []struct {
		name   string
		expRes map[string]string
	}{
		{
			"enabled chains",
			map[string]string{
				utils.ChainWithPublicEndpoints.ChainName:    utils.ChainWithPublicEndpoints.NodeInfo.ChainID,
				utils.ChainWithoutPublicEndpoints.ChainName: utils.ChainWithoutPublicEndpoints.NodeInfo.ChainID,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.ChainIDs()
			s.Require().NoError(err)
			s.Require().NotEmpty(res)
			s.Require().Equal(tt.expRes, res)
		})
	}
}

func (s *TestSuite) TestPrimaryChannelCounterparty() {
	tests := []struct {
		name         string
		chainName    string
		counterparty string
		expRes       cns.ChannelQuery
		success      bool
	}{
		{
			"chain not found",
			"invalidChain",
			"",
			cns.ChannelQuery{},
			false,
		},
		{
			"inserted chain with invalid counterparty name",
			utils.ChainWithPublicEndpoints.ChainName,
			utils.ChainWithoutPublicEndpoints.ChainName,
			cns.ChannelQuery{},
			false,
		},
		{
			"inserted chain with valid counterparty name",
			utils.ChainWithoutPublicEndpoints.ChainName,
			utils.ChainWithPublicEndpoints.ChainName,
			cns.ChannelQuery{
				ChainName:    utils.ChainWithoutPublicEndpoints.ChainName,
				Counterparty: utils.ChainWithPublicEndpoints.ChainName,
				ChannelName:  utils.ChainWithoutPublicEndpoints.PrimaryChannel[utils.ChainWithPublicEndpoints.ChainName],
			},
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.PrimaryChannelCounterparty(tt.chainName, tt.counterparty)
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

func (s *TestSuite) TestPrimaryChannels() {
	tests := []struct {
		name      string
		chainName string
		contains  cns.ChannelQuery
		expLen    int
	}{
		{
			"chain not found",
			"invalidChain",
			cns.ChannelQuery{},
			0,
		},
		{
			"inserted chain",
			utils.ChainWithoutPublicEndpoints.ChainName,
			cns.ChannelQuery{
				ChainName:    utils.ChainWithoutPublicEndpoints.ChainName,
				Counterparty: utils.ChainWithPublicEndpoints.ChainName,
				ChannelName:  utils.ChainWithoutPublicEndpoints.PrimaryChannel[utils.ChainWithPublicEndpoints.ChainName],
			},
			len(utils.ChainWithoutPublicEndpoints.PrimaryChannel),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.PrimaryChannels(tt.chainName)
			s.Require().NoError(err)
			s.Require().Len(res, tt.expLen)
			if tt.expLen != 0 {
				s.Require().NotEmpty(res)
				s.Require().Contains(res, tt.contains)
			}
		})
	}
}

func (s *TestSuite) TestChainsOnlineStatuses() {
	tests := []struct {
		name   string
		expRes []database.ChainOnlineStatusRow
	}{
		{
			"enabled chains",
			[]database.ChainOnlineStatusRow{
				{
					ChainName: utils.ChainWithPublicEndpoints.ChainName,
					Online:    true,
				},
				{
					ChainName: utils.ChainWithoutPublicEndpoints.ChainName,
					Online:    false,
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.ctx.Router.DB.ChainsOnlineStatuses()
			s.Require().NoError(err)
			s.Require().NotEmpty(res)
			s.Require().Equal(tt.expRes, res)
		})
	}
}
