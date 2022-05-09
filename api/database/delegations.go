package database

import (
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/jmoiron/sqlx"
)

// DelegationResponse represents a delegation response got from database.
type DelegationResponse struct {
	tracelistener.DelegationRow

	ValidatorTokens string `db:"tokens" json:"tokens"`
	ValidatorShares string `db:"delegator_shares" json:"delegator_shares"`
}

func (d *Database) Delegations(address string) ([]DelegationResponse, error) {
	var delegations []DelegationResponse

	q, args, err := sqlx.In(`
	SELECT d.chain_name, d.delegator_address, d.validator_address, d.amount, v.tokens, v.delegator_shares
	FROM tracelistener.delegations as d
	INNER JOIN tracelistener.validators as v ON 
		d.validator_address=v.validator_address
	WHERE d.delegator_address IN (?)
	AND d.chain_name IN (
		SELECT chain_name FROM cns.chains WHERE enabled=true
	)
	AND d.delete_height IS NULL
	`, []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return delegations, d.dbi.DB.Select(&delegations, q, args...)
}
