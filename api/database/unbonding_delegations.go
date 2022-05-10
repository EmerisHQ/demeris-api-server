package database

import (
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/jmoiron/sqlx"
)

func (d *Database) UnbondingDelegations(address string) ([]tracelistener.UnbondingDelegationRow, error) {
	var unbondingDelegations []tracelistener.UnbondingDelegationRow

	q, args, err := sqlx.In(`
	SELECT
	id,
	chain_name,
	height,
	delete_height,
	delegator_address,
	validator_address,
	entries
	FROM tracelistener.unbonding_delegations
	WHERE delegator_address IN (?)
	AND chain_name IN (
		SELECT chain_name FROM cns.chains WHERE enabled=true
	)
	AND delete_height IS NULL
	`, []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return unbondingDelegations, d.dbi.DB.Select(&unbondingDelegations, q, args...)
}
