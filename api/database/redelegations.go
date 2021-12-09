package database

import (
	"github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/jmoiron/sqlx"
)

func (d *Database) Redelegations(address string) ([]tracelistener.RedelegationRow, error) {
	var Redelegations []tracelistener.RedelegationRow

	q, args, err := sqlx.In("SELECT * FROM tracelistener.redelegations WHERE delegator_address IN (?) and chain_name in (select chain_name from cns.chains where enabled=true);", []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return Redelegations, d.dbi.DB.Select(&Redelegations, q, args...)
}
