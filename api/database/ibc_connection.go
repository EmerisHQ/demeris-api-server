package database

import (
	"context"

	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/getsentry/sentry-go"
)

func (d *Database) Connection(ctx context.Context, chain string, connection_id string) (tracelistener.IBCConnectionRow, error) {
	defer sentry.StartSpan(ctx, "db.Connection").Finish()

	var connection tracelistener.IBCConnectionRow

	q := `
	SELECT
	id,
	chain_name,
	height,
	delete_height,
	connection_id,
	client_id,
	state,
	counter_connection_id,
	counter_client_id
	FROM tracelistener.connections 
	WHERE chain_name=?
	AND connection_id=?
	AND delete_height IS NULL
	limit 1
	`

	q = d.dbi.DB.Rebind(q)

	if err := d.dbi.DB.GetContext(ctx, &connection, q, chain, connection_id); err != nil {
		return tracelistener.IBCConnectionRow{}, err
	}

	return connection, nil
}
