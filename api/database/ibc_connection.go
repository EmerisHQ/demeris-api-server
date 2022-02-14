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

	return connection, d.dbi.DB.Get(&connection, q, chain, connection_id)
}

func (d *Database) GetConnectionFromChannel(chain, channelId string) (tracelistener.IBCConnectionRow, error) {
	var connection tracelistener.IBCConnectionRow

	q := `
	select * from tracelistener.connections
	where chain_name=?
	and connection_id = (
		select hops[1] from tracelistener.channels
		where channel_id=?
		and   chain_name=?
		and   array_length(hops, 1) = 2
	);
	`
	q = d.dbi.DB.Rebind(q)

	return connection, d.dbi.DB.Get(&connection, q, chain, channelId, chain)
}
