package database_test

import (
	"context"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestUnbondingDelegations() {
	t := s.T()
	ctx := context.Background()
	require := require.New(t)
	assert := assert.New(t)
	err := s.ctx.CnsDB.AddChain(utils.ChainWithoutPublicEndpoints)
	require.NoError(err)
	utils.RunTraceListenerMigrations(s.ctx, t)
	utils.InsertTraceListenerData(s.ctx, t, utils.TracelistenerData{
		UnbondingDelegations: []tracelistener.UnbondingDelegationRow{
			{
				TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
					ChainName: "chain1", Height: 1024,
				},
				Delegator: "dadr1", Validator: "vadr1",
			},
			{
				TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
					ChainName: "chain1", Height: 1024,
				},
				Delegator: "dadr2", Validator: "vadr2",
			},
		},
	})

	// case 1: one address
	bs, err := s.ctx.Router.DB.UnbondingDelegations(ctx, []string{"dadr1"})

	require.NoError(err)
	if assert.Len(bs, 1) {
		assert.Equal("chain1", bs[0].ChainName)
		assert.Equal("dadr1", bs[0].Delegator)
		assert.Equal("vadr1", bs[0].Validator)
		assert.EqualValues(1024, bs[0].Height)
	}

	// case 2: multiple addresses
	bs, err = s.ctx.Router.DB.UnbondingDelegations(ctx, []string{"dadr1", "dadr2"})

	require.NoError(err)
	if assert.Len(bs, 2) {
		assert.Equal("chain1", bs[0].ChainName)
		assert.Equal("dadr1", bs[0].Delegator)
		assert.Equal("vadr1", bs[0].Validator)
		assert.EqualValues(1024, bs[0].Height)

		assert.Equal("chain1", bs[1].ChainName)
		assert.Equal("dadr2", bs[1].Delegator)
		assert.Equal("vadr2", bs[1].Validator)
		assert.EqualValues(1024, bs[1].Height)
	}
}