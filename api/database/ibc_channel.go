package database

import (
	"context"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/getsentry/sentry-go"
)

func (d *Database) GetIbcChannelToChain(ctx context.Context, chain, channel, chainID string) (cns.IbcChannelsInfo, error) {
	defer sentry.StartSpan(ctx, "db.GetIbcChannelToChain").Finish()

	var c cns.IbcChannelsInfo

	subQ := `SELECT
		tracelistener.channels.chain_name,
		tracelistener.channels.channel_id,
		tracelistener.channels.counter_channel_id,
		tracelistener.clients.chain_id
	FROM
		tracelistener.channels
		INNER JOIN tracelistener.connections ON
				tracelistener.channels.hops[1]
				= tracelistener.connections.connection_id
			AND
				tracelistener.connections.chain_name
				= tracelistener.channels.chain_name
		INNER JOIN tracelistener.clients ON
				tracelistener.clients.client_id
				= tracelistener.connections.client_id
			AND
			tracelistener.clients.chain_name
			= tracelistener.channels.chain_name`

	q := `
		SELECT
			c1.chain_name AS chain_a_chain_name,
			c1.channel_id AS chain_a_channel_id,
			c1.counter_channel_id AS chain_a_counter_channel_id,
			c1.chain_id AS chain_a_chain_id,
			c2.chain_name AS chain_b_chain_name,
			c2.channel_id AS chain_b_channel_id,
			c2.counter_channel_id AS chain_b_counter_channel_id,
			c2.chain_id AS chain_b_chain_id
		FROM
			(
				` + subQ + `
			) c1
				INNER	JOIN
			(
				` + subQ + `
			) c2
			ON c1.channel_id = c2.counter_channel_id
			AND c1.counter_channel_id = c2.channel_id

		WHERE
			c1.chain_name != c2.chain_name
			AND c1.chain_name = ?
			AND c1.channel_id = ?
			AND c2.chain_id = ?

		`

	q = d.dbi.DB.Rebind(q)

	err := d.dbi.DB.SelectContext(ctx, &c, q, chain, channel, chainID)
	if err != nil {
		return nil, err
	}

	if len(c) == 0 {
		return nil, ErrNoDestChain{
			Chain_a: chain,
			Channel: channel,
		}
	}

	return c, nil
}
