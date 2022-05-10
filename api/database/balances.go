package database

import (
	"context"

	"github.com/emerishq/demeris-backend-models/tracelistener"
)

func (d *Database) Balances(ctx context.Context, address string) ([]tracelistener.BalanceRow, error) {
	var balances []tracelistener.BalanceRow

	q := `
		SELECT
		id,
		chain_name,
		height,
		delete_height,
		address,
		amount,
		denom
		FROM tracelistener.balances
		WHERE address=?
		AND chain_name IN (
			SELECT chain_name FROM cns.chains WHERE enabled=true
		)
		AND delete_height IS NULL
	`

	q = d.dbi.DB.Rebind(q)

	return balances, d.dbi.DB.SelectContext(ctx, &balances, q, address)
}
