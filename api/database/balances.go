package database

import "github.com/emerishq/demeris-backend-models/tracelistener"

func (d *Database) Balances(address string) ([]tracelistener.BalanceRow, error) {
	var balances []tracelistener.BalanceRow

	q := `
		SELECT * FROM tracelistener.balances
		WHERE address=?
		AND chain_name in (select chain_name from cns.chains where enabled=true)
		AND delete_height IS NULL
	`

	q = d.dbi.DB.Rebind(q)

	return balances, d.dbi.DB.Select(&balances, q, address)
}
