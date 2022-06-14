package database_test

import (
	"context"

	utils "github.com/emerishq/demeris-api-server/api/test_utils"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestDelegations() {
	t := s.T()
	ctx := context.Background()
	require := require.New(t)
	assert := assert.New(t)
	err := s.ctx.CnsDB.AddChain(utils.ChainWithoutPublicEndpoints)
	require.NoError(err)
	utils.RunTraceListenerMigrations(s.ctx, t)
	utils.InsertTraceListenerData(s.ctx, t, utils.TracelistenerData{
		Validators: []tracelistener.ValidatorRow{
			{
				TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
					ChainName: "chain1", Height: 1024,
				},
				OperatorAddress: "opadr", ValidatorAddress: "vadr",
			},
		},
		Delegations: []tracelistener.DelegationRow{
			{
				TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
					ChainName: "chain1", Height: 1024,
				},
				Delegator: "adr1", Validator: "vadr", Amount: "42",
			},
			{
				TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
					ChainName: "chain1", Height: 1024,
				},
				Delegator: "adr2", Validator: "vadr", Amount: "42",
			},
		},
	})

	// case 1: one address
	bs, err := s.ctx.Router.DB.Delegations(ctx, "adr1")

	require.NoError(err)
	if assert.Len(bs, 1) {
		assert.Equal("chain1", bs[0].ChainName)
		assert.Equal("adr1", bs[0].Delegator)
		assert.Equal("vadr", bs[0].Validator)
		assert.Equal("42", bs[0].Amount)
	}

	// case 2: multiple addresses
	bs, err = s.ctx.Router.DB.Delegations(ctx, "adr1", "adr2")

	require.NoError(err)
	if assert.Len(bs, 2) {
		assert.Equal("chain1", bs[0].ChainName)
		assert.Equal("adr1", bs[0].Delegator)
		assert.Equal("vadr", bs[0].Validator)
		assert.Equal("42", bs[0].Amount)

		assert.Equal("chain1", bs[1].ChainName)
		assert.Equal("adr2", bs[1].Delegator)
		assert.Equal("vadr", bs[1].Validator)
		assert.Equal("42", bs[1].Amount)
	}
}
