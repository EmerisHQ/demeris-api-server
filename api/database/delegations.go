package database

import (
	"context"

	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
)

// DelegationResponse represents a delegation response got from database.
type DelegationResponse struct {
	tracelistener.DelegationRow

	ValidatorTokens string `db:"tokens" json:"tokens"`
	ValidatorShares string `db:"delegator_shares" json:"delegator_shares"`
}

func (d *Database) Delegations(ctx context.Context, address string) ([]DelegationResponse, error) {
	defer sentry.StartSpan(ctx, "db.Delegations").Finish()

	var delegations []DelegationResponse

	q, args, err := sqlx.In(`
	SELECT d.chain_name, d.delegator_address, d.validator_address, d.amount, v.tokens, v.delegator_shares
	FROM tracelistener.delegations as d
	INNER JOIN tracelistener.validators as v ON 
		d.validator_address=v.validator_address
	WHERE d.delegator_address=(?)
	AND d.chain_name IN (
		SELECT chain_name FROM cns.chains WHERE enabled=true
	)
	AND v.delete_height IS NULL
	AND d.delete_height IS NULL
	`, []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return delegations, d.dbi.DB.SelectContext(ctx, &delegations, q, args...)
}

func (d *Database) DelegationsOldResponse(ctx context.Context, address string) ([]tracelistener.DelegationRow, error) {
	defer sentry.StartSpan(ctx, "db.DelegationsOldResponse").Finish()

	var delegations []tracelistener.DelegationRow

	q, args, err := sqlx.In(`
	SELECT
	id,
	chain_name,
	height,
	delete_height,
	delegator_address,
	validator_address,
	amount
	FROM tracelistener.delegations
	WHERE delegator_address=(?)
	AND chain_name IN (
		SELECT chain_name FROM cns.chains WHERE enabled=true
	)
	AND delete_height IS NULL
	`, []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return delegations, d.dbi.DB.SelectContext(ctx, &delegations, q, args...)
}
