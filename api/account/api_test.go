package account

import (
	"fmt"
	"testing"

	apimodels "github.com/allinbits/demeris-backend-models/api"
	"github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/stretchr/testify/require"
)

func Test_balanceRespForBalance(t *testing.T) {
	tests := []struct {
		name       string
		rawBalance tracelistener.BalanceRow
		vd         map[string]bool
		dt         denomTraceFunc
		want       apimodels.Balance
	}{
		{
			"verified IBC balance returns verified balance",
			tracelistener.BalanceRow{
				Address: "address",
				Amount:  "42",
				Denom:   "ibc/hash",
			},
			map[string]bool{
				"uatom": true,
			},
			func(_, hash string) (tracelistener.IBCDenomTraceRow, error) {
				return tracelistener.IBCDenomTraceRow{
					Path:      "path",
					BaseDenom: "uatom",
					Hash:      "hash",
				}, nil
			},
			apimodels.Balance{
				Address:   "address",
				BaseDenom: "uatom",
				Verified:  true,
				Amount:    "42",
				OnChain:   "",
				Ibc: apimodels.IbcInfo{
					Path: "path",
					Hash: "hash",
				},
			},
		},
		{
			"non-verified IBC balance returns non-verified balance",
			tracelistener.BalanceRow{
				Address: "address",
				Amount:  "42",
				Denom:   "ibc/hash",
			},
			map[string]bool{
				"uatom": false,
			},
			func(_, hash string) (tracelistener.IBCDenomTraceRow, error) {
				return tracelistener.IBCDenomTraceRow{
					Path:      "path",
					BaseDenom: "uatom",
					Hash:      "hash",
				}, nil
			},
			apimodels.Balance{
				Address:   "address",
				BaseDenom: "uatom",
				Verified:  false,
				Amount:    "42",
				OnChain:   "",
				Ibc: apimodels.IbcInfo{
					Path: "path",
					Hash: "hash",
				},
			},
		},
		{
			"error on denomtrace function returns unverified balance",
			tracelistener.BalanceRow{
				Address: "address",
				Amount:  "42",
				Denom:   "ibc/hash",
			},
			map[string]bool{
				"uatom": true,
			},
			func(_, hash string) (tracelistener.IBCDenomTraceRow, error) {
				return tracelistener.IBCDenomTraceRow{}, fmt.Errorf("error")
			},
			apimodels.Balance{
				Address:   "address",
				BaseDenom: "ibc/hash",
				Verified:  false,
				Amount:    "42",
				OnChain:   "",
				Ibc: apimodels.IbcInfo{
					Hash: "hash",
				},
			},
		},
		{
			"verified non-ibc token returns verified balance",
			tracelistener.BalanceRow{
				Address: "address",
				Amount:  "42",
				Denom:   "denom",
			},
			map[string]bool{
				"denom": true,
			},
			func(_, hash string) (tracelistener.IBCDenomTraceRow, error) {
				return tracelistener.IBCDenomTraceRow{}, nil
			},
			apimodels.Balance{
				Address:   "address",
				BaseDenom: "denom",
				Verified:  true,
				Amount:    "42",
			},
		},
		{
			"non-verified non-ibc token returns non-verified balance",
			tracelistener.BalanceRow{
				Address: "address",
				Amount:  "42",
				Denom:   "denom",
			},
			map[string]bool{
				"denom": false,
			},
			func(_, hash string) (tracelistener.IBCDenomTraceRow, error) {
				return tracelistener.IBCDenomTraceRow{}, nil
			},
			apimodels.Balance{
				Address:   "address",
				BaseDenom: "denom",
				Verified:  false,
				Amount:    "42",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t,
				tt.want,
				balanceRespForBalance(tt.rawBalance, tt.vd, tt.dt),
			)
		})
	}
}
