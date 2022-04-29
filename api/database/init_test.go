package database_test

import (
	"testing"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	ctx *utils.TestingCtx
}

func (s *TestSuite) SetupSuite() {
	// global setup
	s.ctx = utils.Setup(false)

	// insert data
	s.ctx.CnsDB.AddChain(utils.ChainWithPublicEndpoints)
	s.ctx.CnsDB.AddChain(utils.ChainWithoutPublicEndpoints)
	s.ctx.CnsDB.AddChain(utils.DisabledChain)

	utils.RunTraceListnerMigrations(s.ctx, s.T())
	utils.InsertTraceListnerData(s.ctx, s.T(), utils.VerifyTraceData)
}

func (s *TestSuite) TearDownSuite() {
	s.T().Log("tearing down database test suite")
	s.Require().NoError(s.ctx.Router.DB.Close())
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
