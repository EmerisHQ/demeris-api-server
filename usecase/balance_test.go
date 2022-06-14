package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/emerishq/demeris-api-server/api/account"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalances(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name             string
		addresses        []string
		expectedError    string
		expectedBalances []account.Balance
		setup            func(mocks)
	}{
		{
			name:             "ok: empty addresses",
			expectedBalances: []account.Balance{},
		},
		{
			name:          "fail: adresses not found",
			addresses:     []string{"adr1"},
			expectedError: "balances not found for addresses [adr1]",

			setup: func(m mocks) {
				m.db.EXPECT().Balances(ctx, []string{"adr1"}).Return(nil, nil)
			},
		},
		{
			name:      "ok: denom unverified",
			addresses: []string{"adr1"},
			expectedBalances: []account.Balance{
				{
					Address:   "adr1",
					BaseDenom: "denom1",
					Amount:    "42",
				},
			},

			setup: func(m mocks) {
				m.db.EXPECT().Balances(ctx, []string{"adr1"}).Return(
					[]tracelistener.BalanceRow{
						{Address: "adr1", Denom: "denom1", Amount: "42"},
					},
					nil,
				)
				m.db.EXPECT().VerifiedDenoms(ctx).Return(map[string]cns.DenomList{}, nil)
			},
		},
		{
			name:      "ok: denom verified",
			addresses: []string{"adr1"},
			expectedBalances: []account.Balance{
				{
					Address:   "adr1",
					BaseDenom: "denom1",
					Amount:    "42",
					Verified:  true,
				},
			},

			setup: func(m mocks) {
				m.db.EXPECT().Balances(ctx, []string{"adr1"}).Return(
					[]tracelistener.BalanceRow{
						{Address: "adr1", Denom: "denom1", Amount: "42"},
					},
					nil,
				)
				m.db.EXPECT().VerifiedDenoms(ctx).Return(map[string]cns.DenomList{
					"xxx": cns.DenomList{
						{Name: "denom1", Verified: true},
					},
				}, nil)
			},
		},
		{
			name:      "ok: unverified ibc denom from chain2",
			addresses: []string{"adr1"},
			expectedBalances: []account.Balance{
				{
					Address:   "adr1",
					BaseDenom: "denom2",
					Amount:    "42",
					OnChain:   "chain2",
					Ibc: account.IbcInfo{
						Path: "path",
						Hash: "xxx",
					},
				},
			},

			setup: func(m mocks) {
				m.db.EXPECT().Balances(ctx, []string{"adr1"}).Return(
					[]tracelistener.BalanceRow{
						{
							TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
								ChainName: "chain2",
							},
							Address: "adr1",
							Denom:   "ibc/xxx",
							Amount:  "42",
						},
					},
					nil,
				)
				m.db.EXPECT().VerifiedDenoms(ctx).Return(map[string]cns.DenomList{
					"xxx": cns.DenomList{
						{Name: "denom1", Verified: true},
						{Name: "denom2", Verified: false},
					}}, nil)
				m.db.EXPECT().DenomTrace(ctx, "chain2", "xxx").Return(
					tracelistener.IBCDenomTraceRow{
						BaseDenom: "denom2",
						Path:      "path",
					},
					nil,
				)
			},
		},
		{
			name:      "ok: verified ibc denom from chain2",
			addresses: []string{"adr1"},
			expectedBalances: []account.Balance{
				{
					Address:   "adr1",
					BaseDenom: "denom2",
					Amount:    "42",
					OnChain:   "chain2",
					Verified:  true,
					Ibc: account.IbcInfo{
						Path: "path",
						Hash: "xxx",
					},
				},
			},

			setup: func(m mocks) {
				m.db.EXPECT().Balances(ctx, []string{"adr1"}).Return(
					[]tracelistener.BalanceRow{
						{
							TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
								ChainName: "chain2",
							},
							Address: "adr1",
							Denom:   "ibc/xxx",
							Amount:  "42",
						},
					},
					nil,
				)
				m.db.EXPECT().VerifiedDenoms(ctx).Return(map[string]cns.DenomList{
					"xxx": cns.DenomList{
						{Name: "denom2", Verified: true},
					}}, nil)
				m.db.EXPECT().DenomTrace(ctx, "chain2", "xxx").Return(
					tracelistener.IBCDenomTraceRow{
						BaseDenom: "denom2",
						Path:      "path",
					},
					nil,
				)
			},
		},
		{
			name:      "ok: DenomTrace returns an error",
			addresses: []string{"adr1"},
			expectedBalances: []account.Balance{
				{
					Address:   "adr1",
					BaseDenom: "ibc/xxx",
					Amount:    "42",
					OnChain:   "chain2",
					Ibc: account.IbcInfo{
						Path: "",
						Hash: "xxx",
					},
				},
			},

			setup: func(m mocks) {
				m.db.EXPECT().Balances(ctx, []string{"adr1"}).Return(
					[]tracelistener.BalanceRow{
						{
							TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
								ChainName: "chain2",
							},
							Address: "adr1",
							Denom:   "ibc/xxx",
							Amount:  "42",
						},
					},
					nil,
				)
				m.db.EXPECT().VerifiedDenoms(ctx).Return(map[string]cns.DenomList{
					"xxx": cns.DenomList{
						{Name: "denom1", Verified: true},
						{Name: "denom2", Verified: true},
					}}, nil)
				m.db.EXPECT().DenomTrace(ctx, "chain2", "xxx").
					Return(tracelistener.IBCDenomTraceRow{}, errors.New("oups"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newApp(t, tt.setup)

			balances, err := app.Balances(ctx, tt.addresses)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBalances, balances)
		})
	}
}
