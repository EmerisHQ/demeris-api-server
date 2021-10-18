package database

import "github.com/allinbits/demeris-backend-models/tracelistener"

func (d *Database) Connection(chain string, connection_id string) (tracelistener.IBCConnectionRow, error) {
	var connection tracelistener.IBCConnectionRow

	q := `
	SELECT *
	FROM tracelistener.connections 
	WHERE chain_name=? AND connection_id=?;
	`

	q = d.dbi.DB.Rebind(q)

	return connection, d.dbi.DB.Select(&connection, q, chain, connection_id)
}
