package database

import (
	"github.com/emerishq/demeris-backend-models/tracelistener"
)

// DenomTrace returns the denom trace for a given chain by its hash. Hash param is case-insensitive.
func (d *Database) DenomTrace(chain string, hash string) (tracelistener.IBCDenomTraceRow, error) {
	var denomTrace tracelistener.IBCDenomTraceRow

	// note: lower() since Tracelistener stores hashes in lowercase
	q := `
	SELECT
	id,
	chain_name,
	height,
	delete_height,
	path,
	base_denom,
	hash
	FROM tracelistener.denom_traces
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
