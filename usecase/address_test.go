package usecase_test

import (
	"context"
	"testing"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveRawAddress(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name              string
		rawAddress        string
		expectedError     string
		expectedAddresses []string

		setup func(mocks)
	}{
		{
			name:          "fail: empty raw address",
			rawAddress:    "",
			expectedError: "raw address is empty",
		},
		{
			name:          "fail: raw address is not in hex format",
			rawAddress:    "-",
			expectedError: "raw address is not in hex format: encoding/hex: invalid byte: U+002D '-'",
		},
		{
			name:              "ok: no chain enabled",
			rawAddress:        "abc123",
			expectedAddresses: []string{},

			setup: func(m mocks) {
				m.db.EXPECT().Chains(ctx).Return([]cns.Chain{}, nil)
			},
		},
		{
			name:       "ok",
			rawAddress: "abc123",
			expectedAddresses: []string{
				"pre1140qjx2fqawy",
				"pre2140qjxcavcur",
			},

			setup: func(m mocks) {
				m.db.EXPECT().Chains(ctx).Return([]cns.Chain{
					{
						ChainName: "chain1",
						NodeInfo: cns.NodeInfo{
							Bech32Config: cns.Bech32Config{MainPrefix: "pre1"},
						},
					},
					{
						ChainName: "chain2",
						NodeInfo: cns.NodeInfo{
							Bech32Config: cns.Bech32Config{MainPrefix: "pre2"},
						},
					},
				},
					nil,
				)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newApp(t, tt.setup)

			adrs, err := app.DeriveRawAddress(ctx, tt.rawAddress)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAddresses, adrs)
		})
	}
}
