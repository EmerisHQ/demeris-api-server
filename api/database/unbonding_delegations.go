package database

import (
	"github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/jmoiron/sqlx"
)

func (d *Database) UnbondingDelegations(address string) ([]tracelistener.UnbondingDelegationRow, error) {
	var unbondingDelegations []tracelistener.UnbondingDelegationRow

	q, args, err := sqlx.In("SELECT * FROM tracelistener.unbonding_delegations WHERE delegator_address IN (?) and chain_name in (select chain_name from cns.chains where enabled=true);", []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return unbondingDelegations, d.dbi.DB.Select(&unbondingDelegations, q, args...)
}
