package database

import (
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/jmoiron/sqlx"
)

func (d *Database) Numbers(address string) ([]tracelistener.AuthRow, error) {
	var numbers []tracelistener.AuthRow

	q, args, err := sqlx.In(`
	SELECT 
		id,
		chain_name,
		height,
		delete_height,
		address,
		sequence_number,
		account_number
	FROM tracelistener.auth
	WHERE address IN (?)
	AND chain_name IN (
		SELECT chain_name FROM cns.chains WHERE enabled=true
	)
	AND delete_height IS NULL
	`, []string{address})
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
