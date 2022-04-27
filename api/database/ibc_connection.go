package database

import (
	"fmt"

	"github.com/emerishq/demeris-backend-models/tracelistener"
)

func (d *Database) Connection(chain string, connection_id string) (tracelistener.IBCConnectionRow, error) {
	var connections []tracelistener.IBCConnectionRow

	q := `
	SELECT *
	FROM tracelistener.connections 
	WHERE chain_name=?
	AND connection_id=?
	AND delete_height IS NULL
	limit 1
	`

	q = d.dbi.DB.Rebind(q)

	if err := d.dbi.DB.Select(&connections, q, chain, connection_id); err != nil {
		return tracelistener.IBCConnectionRow{}, err
	}

	if len(connections) == 0 {
		return tracelistener.IBCConnectionRow{}, fmt.Errorf("query done but returned no result")
	}

	return connections[0], nil
}
