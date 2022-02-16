package chains_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	createTraceListenerDatabase = `
	CREATE DATABASE IF NOT EXISTS tracelistener;
	`
	createDenomTracesTable = `
	CREATE TABLE IF NOT EXISTS tracelistener.denom_traces (
		id serial unique primary key,
		chain_name text not null,
		path text not null,
		base_denom text not null,
		hash text not null,
		unique(chain_name, hash)
	)
	`
	createChannelsTable = `
	CREATE TABLE IF NOT EXISTS tracelistener.channels (
		id serial unique primary key,
		chain_name text not null,
		channel_id text not null,
		counter_channel_id text not null,
		port text not null,
		state integer not null,
		hops text[] not null,
		unique(chain_name, channel_id, port)
	)
	`
	createConnectionsTable = `
	CREATE TABLE IF NOT EXISTS tracelistener.connections (
		id serial unique primary key,
		chain_name text not null,
		connection_id text not null,
		client_id text not null,
		state text not null,
		counter_connection_id text not null,
		counter_client_id text not null,
		unique(chain_name, connection_id, client_id)
	)
	`

	createClientsTable = `
	CREATE TABLE IF NOT EXISTS tracelistener.clients (
		id serial unique primary key,
		chain_name text not null,
		chain_id text not null,
		client_id text not null,
		latest_height numeric not null,
		trusting_period numeric not null,
		unique(chain_name, chain_id, client_id)
	)
	`
	createBlockTimeTable = `CREATE TABLE IF NOT EXISTS tracelistener.blocktime (
		id serial unique primary key,
		chain_name text not null,
		block_time timestamp not null,
		unique(chain_name)
	)`

	insertDenomTrace = "INSERT INTO tracelistener.denom_traces (path, base_denom, hash, chain_name) VALUES (($1), ($2), ($3), ($4)) ON CONFLICT (chain_name, hash) DO UPDATE SET base_denom=($2), hash=($3), path=($1)"
	insertChannel    = "INSERT INTO tracelistener.channels (channel_id, counter_channel_id, port, state, hops, chain_name) VALUES (($1), ($2), ($3), ($4), ($5), ($6)) ON CONFLICT (chain_name, channel_id, port) DO UPDATE SET state=($4),counter_channel_id=($2),hops=($5),port=($3),channel_id=($1)"
	insertConnection = "INSERT INTO tracelistener.connections (chain_name, connection_id, client_id, state, counter_connection_id, counter_client_id) VALUES (($1), ($2), ($3), ($4), ($5), ($6)) ON CONFLICT (chain_name, connection_id, client_id) DO UPDATE SET chain_name=($1),state=($4),counter_connection_id=($5),counter_client_id=($6)"
	insertClient     = "INSERT INTO tracelistener.clients (chain_name, chain_id, client_id, latest_height, trusting_period) VALUES (($1), ($2), ($3), ($4), ($5)) ON CONFLICT (chain_name, chain_id, client_id) DO UPDATE SET chain_id=($2),client_id=($3),latest_height=($4),trusting_period=($5)"
	insertBlocktime  = "INSERT INTO tracelistener.blocktime (chain_name, block_time) VALUES (($1), ($2)) ON CONFLICT (chain_name) DO UPDATE SET chain_name=($1),block_time=($2);"
)

var migrations = []string{
	createTraceListenerDatabase,
	createDenomTracesTable,
	createChannelsTable,
	createConnectionsTable,
	createClientsTable,
	createBlockTimeTable,
}

func runTraceListnerMigrations(t *testing.T) {
	for _, m := range migrations {
		_, err := testingCtx.CnsDB.Instance.DB.Exec(m)
		require.NoError(t, err)
	}
}

func insertRow(t *testing.T, query string, args ...interface{}) {

	res, err := testingCtx.CnsDB.Instance.DB.Exec(query, args...)
	require.NoError(t, err)

	rows, _ := res.RowsAffected()

	require.NotEqual(t, 0, rows)
}

func insertTraceListnerData(t *testing.T, data tracelistenerData) {
	for _, d := range data.denoms {
		insertRow(t, insertDenomTrace, d.path, d.baseDenom, d.hash, d.chainName)
	}

	for _, d := range data.channels {
		insertRow(t, insertChannel, d.channelID, d.counterChannelID, d.port, d.state, d.hops, d.chainName)
	}

	for _, d := range data.connections {
		insertRow(t, insertConnection, d.chainName, d.connectionID, d.clientID, d.state, d.counterConnectionID, d.counterClientID)
	}

	for _, d := range data.clients {
		insertRow(t, insertClient, d.sourceChainName, d.destChainID, d.clientID, d.latestHeight, d.trustingPeriod)
	}

	for _, d := range data.blockTimes {
		insertRow(t, insertBlocktime, d.chainName, d.time)
	}
}
