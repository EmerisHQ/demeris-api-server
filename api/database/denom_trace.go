package database

import (
	"fmt"

	"github.com/emerishq/demeris-backend-models/tracelistener"
)

// DenomTrace returns the denom trace for a given chain by its hash. Hash param is case-insensitive.
func (d *Database) DenomTrace(chain string, hash string) (tracelistener.IBCDenomTraceRow, error) {
	var denomTrace tracelistener.IBCDenomTraceRow

	// note: lower() since Tracelistener stores hashes in lowercase
	q := `
	SELECT * FROM tracelistener.denom_traces
	WHERE chain_name=?
	AND hash=lower(?)
	AND base_denom != ''
	AND delete_height IS NULL
	LIMIT 1
	`

	q = d.dbi.DB.Rebind(q)

	if err := d.dbi.DB.Get(&denomTrace, q, chain, hash); err != nil {
		return tracelistener.IBCDenomTraceRow{}, err
	}

	return denomTrace, nil
}

// DenomTraces returns the all the denom trace for a given chain. Hash param is case-insensitive.
func (d *Database) DenomTraces(chain string, hash string) ([]tracelistener.IBCDenomTraceRow, error) {
	var denomTraces []tracelistener.IBCDenomTraceRow

	// note: lower() since Tracelistener stores hashes in lowercase
	q := `
	SELECT * FROM tracelistener.denom_traces
	WHERE chain_name=?
	AND base_denom != ''
	AND delete_height IS NULL
	`

	q = d.dbi.DB.Rebind(q)

	if err := d.dbi.DB.Select(&denomTraces, q, chain); err != nil {
		return []tracelistener.IBCDenomTraceRow{}, err
	}

	if len(denomTraces) == 0 {
		return []tracelistener.IBCDenomTraceRow{}, fmt.Errorf("query done but returned no result")
	}

	return denomTraces, nil
}
