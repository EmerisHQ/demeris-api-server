package database

import (
	"context"

	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
)

func (d *Database) Balances(ctx context.Context, addresses []string) ([]tracelistener.BalanceRow, error) {
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
		WHERE address IN (?)
		AND chain_name IN (
			SELECT chain_name FROM cns.chains WHERE enabled=true
		)
		AND delete_height IS NULL`
	q, args, err := sqlx.In(q, addresses)
	if err != nil {
		return nil, err
	}
	q = d.dbi.DB.Rebind(q)
	return balances, d.dbi.DB.SelectContext(ctx, &balances, q, args...)
}
