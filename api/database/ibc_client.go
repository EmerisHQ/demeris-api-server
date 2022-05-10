package database

import (
	"context"
	"fmt"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/lib/pq"
)

func (d *Database) QueryIBCClientTrace(ctx context.Context, chain string, channel string) ([]cns.IbcClientInfo, error) {
	clients := []cns.IbcClientInfo{}

	q := `
	SELECT 
		conn.chain_name as chain_name, 
		conn.connection_id as connection_id,
		conn.client_id as client_id,
		ch.channel_id as channel_id, 
		conn.counter_connection_id as counter_connection_id,
		conn.counter_client_id as counter_client_id,
		ch.port as port, 
		ch.state as state, 
		ch.hops as hops 
	FROM tracelistener.connections conn 
	INNER JOIN 
		(SELECT
			id,
			chain_name,
			height,
			delete_height,
			channel_id,
			counter_channel_id,
			hops,
			port,
			state
			FROM tracelistener.channels 
			WHERE chain_name=:chain_name
			AND channel_id=:channel_id
			AND delete_height IS NULL
		) ch 
	ON conn.connection_id=ANY(ch.hops)
	WHERE conn.delete_height IS NULL;
	`
	rows, err := d.dbi.DB.NamedQueryContext(ctx, q, map[string]interface{}{
		"chain_name": chain,
		"channel_id": channel,
	})

	if err != nil {
		return clients, err
	}

	for rows.Next() {
		var client cns.IbcClientInfo
		err = rows.Scan(&client.ChainName, &client.ConnectionId, &client.ClientId, &client.ChannelId, &client.CounterConnectionID,
			&client.CounterClientID, &client.Port, &client.State, pq.Array(&client.Hops))
		if err != nil {
			return clients, err
		}

		clients = append(clients, client)
	}

	err = rows.Close()
	if err != nil {
		return clients, fmt.Errorf("closing rows object: %w", err)
	}

	if len(clients) == 0 {
		return []cns.IbcClientInfo{}, fmt.Errorf("query done but returned no result")
	}

	return clients, nil
}
