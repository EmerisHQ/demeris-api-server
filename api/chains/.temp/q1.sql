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
base_denom as (
    select
        *
    from
        (
            select
                chain_name,
                jsonb_array_elements(denoms) ->> 'name' as name,
                jsonb_array_elements(denoms) ->> 'fee_token' as fee_token,
                jsonb_array_elements(denoms) ->> 'stakable' as stakable
            from
                cns.chains
        )
    where
        fee_token = 'true'
)
select
    *
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
    inner join base_denom on ibc_token.chain_name = base_denom.chain_name
    and ibc_token.base_denom = base_denom.name;