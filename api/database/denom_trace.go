package database

import (
	"fmt"
	"strings"

	"github.com/emerishq/demeris-backend-models/tracelistener"
)

func (d *Database) DenomTrace(chain string, hash string) (tracelistener.IBCDenomTraceRow, error) {
	hash = strings.ToLower(hash)
	var denomTraces []tracelistener.IBCDenomTraceRow

	q := `
	SELECT * FROM tracelistener.denom_traces
	WHERE chain_name=?
	AND hash=?
	AND base_denom != ''
	AND delete_height IS NULL
	LIMIT 1
	`

	q = d.dbi.DB.Rebind(q)

	if err := d.dbi.DB.Select(&denomTraces, q, chain, hash); err != nil {
		return tracelistener.IBCDenomTraceRow{}, err
	}

	if len(denomTraces) == 0 {
		return tracelistener.IBCDenomTraceRow{}, fmt.Errorf("query done but returned no result")
	}

	return denomTraces[0], nil
}
