package chains_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/allinbits/demeris-api-server/api/chains"
	utils "github.com/allinbits/demeris-api-server/api/test_utils"

	"github.com/allinbits/demeris-backend-models/cns"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

const (
	chainEndpointUrl       = "http://%s/chain/%s"
	chainsEndpointUrl      = "http://%s/chains"
	verifyTraceEndpointUrl = "http://%s/chain/%s/denom/verify_trace/%s"
)

func TestGetChain(t *testing.T) {

	tests := []struct {
		name             string
		dataStruct       cns.Chain
		chainName        string
		expectedHttpCode int
		success          bool
	}{
		{
			"Get Chain - Unknown chain",
			cns.Chain{}, // ignored
			"foo",
			400,
			false,
		},
		{
			"Get Chain - Without PublicEndpoint",
			chainWithoutPublicEndpoints,
			chainWithoutPublicEndpoints.ChainName,
			200,
			true,
		},
		{
			"Get Chain - With PublicEndpoints",
			chainWithPublicEndpoints,
			chainWithPublicEndpoints.ChainName,
			200,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// arrange
			// if we have a populated Chain store, add it
			if !cmp.Equal(tt.dataStruct, cns.Chain{}) {
				err := testingCtx.CnsDB.AddChain(tt.dataStruct)
				require.NoError(t, err)
			}

			// act
			resp, err := http.Get(fmt.Sprintf(chainEndpointUrl, testingCtx.Cfg.ListenAddr, tt.chainName))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// assert
			if !tt.success {
				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			respStruct := chains.ChainResponse{}
			err = json.Unmarshal(body, &respStruct)
			require.NoError(t, err)

			require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
			require.Equal(t, tt.dataStruct, respStruct.Chain)
		})
	}
	utils.TruncateDB(testingCtx, t)
}

func TestGetChains(t *testing.T) {

	tests := []struct {
		name             string
		dataStruct       []cns.Chain
		expectedHttpCode int
		success          bool
	}{
		{
			"Get Chains - Nothing in DB",
			[]cns.Chain{}, // ignored
			200,
			true,
		},
		{
			"Get Chains - 2 Chains (With & Without)",
			[]cns.Chain{chainWithoutPublicEndpoints, chainWithPublicEndpoints},
			200,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// arrange
			// if we have a populated Chain store, add it
			if len(tt.dataStruct) != 0 {
				for _, c := range tt.dataStruct {
					err := testingCtx.CnsDB.AddChain(c)
					require.NoError(t, err)
				}
			}

			// act
			resp, err := http.Get(fmt.Sprintf(chainsEndpointUrl, testingCtx.Cfg.ListenAddr))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// assert
			if !tt.success {
				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				respStruct := chains.ChainsResponse{}
				err = json.Unmarshal(body, &respStruct)
				require.NoError(t, err)

				require.Equal(t, tt.expectedHttpCode, resp.StatusCode)
				for _, c := range tt.dataStruct {
					require.Contains(t, respStruct.Chains, toSupportedChain(c))
				}
			}
		})
	}
	utils.TruncateDB(testingCtx, t)
}

func TestVerifyTrace(t *testing.T) {
	runMigrations(t)
	insertRow(t, insertDenomTrace, "transfer/ch1", "denom2", "12345", "chain1")
	insertRow(t, insertChannel, "ch1", "ch2", "p1", 3, []string{"conn1"}, "chain1")
	insertRow(t, insertChannel, "ch2", "ch1", "p2", 3, []string{"conn2"}, "chain2")
	insertRow(t, insertConnection, "chain1", "conn1", "cl1", "ready", "conn2", "cl2")
	insertRow(t, insertConnection, "chain2", "conn2", "cl2", "ready", "conn1", "cl1")
	insertRow(t, insertClient, "chain1", "chain_2", "cl1", "99", "10")
	insertRow(t, insertClient, "chain2", "chain_1", "cl2", "99", "10")
	insertRow(t, insertBlocktime, "chain2", time.Now())

	testingCtx.CnsDB.AddChain(chainWithoutPublicEndpoints)
	testingCtx.CnsDB.AddChain(chainWithPublicEndpoints)
	// act
	resp, err := http.Get(fmt.Sprintf(verifyTraceEndpointUrl, testingCtx.Cfg.ListenAddr, "chain1", "12345"))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var data map[string]interface{}
	err = json.Unmarshal(b, &data)
	require.NoError(t, err)

	fmt.Println(data)
	require.NotEqual(t, 0, 0)
}

func toSupportedChain(c cns.Chain) chains.SupportedChain {

	return chains.SupportedChain{
		ChainName:   c.ChainName,
		DisplayName: c.DisplayName,
		Logo:        c.Logo,
	}
}

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

func runMigrations(t *testing.T) {
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
