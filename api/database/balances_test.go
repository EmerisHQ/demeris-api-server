package database_test

import (
	"context"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestBalances() {
	t := s.T()
	ctx := context.Background()
	require := require.New(t)
	assert := assert.New(t)
	err := s.ctx.CnsDB.AddChain(utils.ChainWithoutPublicEndpoints)
	require.NoError(err)
	utils.RunTraceListenerMigrations(s.ctx, t)
	utils.InsertTraceListenerData(s.ctx, t, utils.TracelistenerData{
		Balances: []tracelistener.BalanceRow{
			{
				TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
					ChainName: "chain1", Height: 1024,
				},
				Address: "adr1", Amount: "42", Denom: "denom1",
			},
			{
				TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
					ChainName: "chain1", Height: 1024,
				},
				Address: "adr2", Amount: "42", Denom: "denom2",
			},
		},
	})

	// case 1: one address
	bs, err := s.ctx.Router.DB.Balances(ctx, []string{"adr1"})

	require.NoError(err)
	if assert.Len(bs, 1) {
		assert.Equal("chain1", bs[0].ChainName)
		assert.Equal("denom1", bs[0].Denom)
		assert.Equal("adr1", bs[0].Address)
		assert.Equal("42", bs[0].Amount)
		assert.EqualValues(1024, bs[0].Height)
	}

	// case 2: multiple addresses
	bs, err = s.ctx.Router.DB.Balances(ctx, []string{"adr1", "adr2"})

	require.NoError(err)
	if assert.Len(bs, 2) {
		assert.Equal("chain1", bs[0].ChainName)
		assert.Equal("denom1", bs[0].Denom)
		assert.Equal("adr1", bs[0].Address)
		assert.Equal("42", bs[0].Amount)
		assert.EqualValues(1024, bs[0].Height)

		assert.Equal("chain1", bs[1].ChainName)
		assert.Equal("denom2", bs[1].Denom)
		assert.Equal("adr2", bs[1].Address)
		assert.Equal("42", bs[1].Amount)
		assert.EqualValues(1024, bs[1].Height)
	}
}
