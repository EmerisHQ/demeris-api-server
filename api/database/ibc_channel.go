package database

import (
	"github.com/emerishq/demeris-backend-models/cns"
)

func (d *Database) GetIbcChannelToChain(chain, channel, chainID string) (cns.IbcChannelsInfo, error) {
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

	err := d.dbi.DB.Select(&c, q, chain, channel, chainID)
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

func (d *Database) GetChannelMatchingDenoms() (ChannelConnectionMatchingDenoms, error) {
	var c ChannelConnectionMatchingDenoms

	q := `
	with ibc_token as (
		select
			chain_name,
			path,
			hash,
			base_denom
		from
			tracelistener.denom_traces
		where
			path like 'transfer/%'
	),
	base_denoms as (
		select
			chain_name,
			base_token,
			fee_token,
			stakable
		from
			(
				select
					chain_name,
					jsonb_array_elements(denoms) ->> 'name' as base_token,
					jsonb_array_elements(denoms) ->> 'fee_token' as fee_token,
					jsonb_array_elements(denoms) ->> 'stakable' as stakable
				from
					cns.chains
				where
					enabled = TRUE
			)
		where
			fee_token = 'true'
	)
	select distinct
		ibc_token.chain_name as chain_name,
		channel_info.channel_id as channel_id,
		channel_info.from_chain as counterparty_chain,
		channel_info.counterparty_channel_id as counterparty_channel_id,
		ibc_token.base_denom as base_denom,
		ibc_token.hash as hash
	from
		ibc_token
		inner join (
			select
				src.chain_name as on_chain,
				dest_channel.chain_name as from_chain,
				src.connection_id as connection_id,
				src.counter_connection_id as counterparty_connection_id,
				src.client_id as client_id,
				src.counter_client_id as counterparty_client_id,
				src_channel.channel_id as channel_id,
				src_channel.counter_channel_id as counterparty_channel_id,
				src_channel.port as port_id,
				src.state as state,
				concat(src_channel.port, '/', src_channel.channel_id) as path
			from
				tracelistener.connections as src
				inner join tracelistener.connections as dest on src.client_id = dest.counter_client_id
				and src.connection_id = dest.counter_connection_id
				and src.counter_client_id = dest.client_id
				and src.counter_connection_id = dest.connection_id
				and src.state = 'STATE_OPEN'
				and dest.state = 'STATE_OPEN'
				inner join (
					select
						chain_name,
						channel_id,
						counter_channel_id,
						port,
						state,
						unnest(hops) as connection_id
					from
						tracelistener.channels
					where
						port = 'transfer'
						and state = 3
				) as src_channel on src.connection_id = src_channel.connection_id
				and src.chain_name = src_channel.chain_name
				inner join (
					select
						chain_name,
						channel_id,
						counter_channel_id,
						port,
						state,
						unnest(hops) as connection_id
					from
						tracelistener.channels
					where
						port = 'transfer'
						and state = 3
				) as dest_channel on src.counter_connection_id = dest_channel.connection_id
				and src_channel.counter_channel_id = dest_channel.channel_id
				and src_channel.channel_id = dest_channel.counter_channel_id
		) as channel_info on channel_info.on_chain = ibc_token.chain_name
		and channel_info.path = ibc_token.path
		inner join base_denoms on base_denoms.chain_name = channel_info.from_chain
		and base_denoms.base_token = ibc_token.base_denom
	order by
		ibc_token.chain_name,
		channel_info.from_chain
	`

	q = d.dbi.DB.Rebind(q)

	err := d.dbi.DB.Select(&c, q)
	if err != nil {
		return nil, err
	}

	if len(c) == 0 {
		return nil, ErrNoMatchingChannels{}
	}

	return c, nil
}
