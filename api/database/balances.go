package database

import (
	"context"

	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/getsentry/sentry-go"
)

func (d *Database) Balances(ctx context.Context, address string) ([]tracelistener.BalanceRow, error) {
	defer sentry.StartSpan(ctx, "db.Balances").Finish()

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
