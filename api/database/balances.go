package database

import "github.com/emerishq/demeris-backend-models/tracelistener"

func (d *Database) Balances(address string) ([]tracelistener.BalanceRow, error) {
	var balances []tracelistener.BalanceRow

	q := `
		SELECT * FROM tracelistener.balances
		WHERE address=?
		AND chain_name IN (
			SELECT chain_name FROM cns.chains WHERE enabled=true
		)
		AND delete_height IS NULL
	`

	q = d.dbi.DB.Rebind(q)

	return balances, d.dbi.DB.Select(&balances, q, address)
}
