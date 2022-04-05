package database

import "github.com/emerishq/demeris-backend-models/tracelistener"

func (d *Database) GetValidators(chain string) ([]tracelistener.ValidatorRow, error) {
	var validators []tracelistener.ValidatorRow

	q := `
	SELECT *
	FROM tracelistener.validators 
	WHERE chain_name=?
	AND delete_height IS NULL
	`

	q = d.dbi.DB.Rebind(q)

	return validators, d.dbi.DB.Select(&validators, q, chain)
}
