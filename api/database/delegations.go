package database

import (
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/jmoiron/sqlx"
)

func (d *Database) Delegations(address string) ([]tracelistener.DelegationRow, error) {
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

	return delegations, d.dbi.DB.Select(&delegations, q, args...)
}
