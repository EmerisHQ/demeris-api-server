package database

import (
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/jmoiron/sqlx"
)

func (d *Database) Numbers(address string) ([]tracelistener.AuthRow, error) {
	var numbers []tracelistener.AuthRow

	q, args, err := sqlx.In("SELECT * FROM tracelistener.auth WHERE address IN (?) and chain_name in (select chain_name from cns.chains where enabled=true);", []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return numbers, d.dbi.DB.Select(&numbers, q, args...)
}

type ChainName struct {
	ChainName     string `db:"chain_name"`
	AccountPrefix string `db:"account_prefix"`
}

func (d *Database) ChainNames() ([]ChainName, error) {
	var cn []ChainName

	q := `select chain_name,node_info->'bech32_config'->>'prefix_account' as account_prefix from cns.chains where enabled=true`

	q = d.dbi.DB.Rebind(q)

	return cn, d.dbi.DB.Select(&cn, q)
}
