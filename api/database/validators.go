package database

import "github.com/allinbits/demeris-backend-models/tracelistener"

func (d *Database) GetValidators(chain string) ([]tracelistener.ValidatorRow, error) {
	var validators []tracelistener.ValidatorRow

	q := `
	SELECT *
	FROM tracelistener.validators 
	WHERE chain_name=?;
	`

	q = d.dbi.DB.Rebind(q)

	return validators, d.dbi.DB.Select(&validators, q, chain)
}
