package database

import (
	"context"

	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/getsentry/sentry-go"
)

func (d *Database) GetValidators(ctx context.Context, chain string) ([]tracelistener.ValidatorRow, error) {
	defer sentry.StartSpan(ctx, "db.GetValidators").Finish()

	var validators []tracelistener.ValidatorRow

	q := `
	SELECT
	id,
	chain_name,
	height,
	delete_height,
	operator_address,
	consensus_pubkey_type,
	consensus_pubkey_value,
	jailed,
	status,
	tokens,
	delegator_shares,
	moniker,
	identity,
	website,
	security_contact,
	details,
	unbonding_height,
	unbonding_time,
	commission_rate,
	max_rate,
	max_change_rate,
	update_time,
	min_self_delegation
	FROM tracelistener.validators 
	WHERE chain_name=?
	AND delete_height IS NULL
	`

	q = d.dbi.DB.Rebind(q)

	return validators, d.dbi.DB.SelectContext(ctx, &validators, q, chain)
}
