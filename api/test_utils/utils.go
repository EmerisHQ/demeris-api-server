package test_utils

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/emerishq/emeris-utils/store"
	"github.com/stretchr/testify/require"

	"github.com/emerishq/demeris-api-server/api/config"
	apiDb "github.com/emerishq/demeris-api-server/api/database"
	cnsDb "github.com/emerishq/emeris-cns-server/cns/database"

	"github.com/alicebob/miniredis/v2"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/emerishq/demeris-api-server/api/router"
	"github.com/emerishq/demeris-api-server/mocks"
	"github.com/emerishq/emeris-utils/logging"
	"go.uber.org/zap"
)

const k8sNsInTest = "emeris"

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

	truncateDenomTraces = `TRUNCATE tracelistener.denom_traces`
	truncateChannels    = `TRUNCATE tracelistener.channels`
	truncateConnections = `TRUNCATE tracelistener.connections`
	truncateClients     = `TRUNCATE tracelistener.clients`
	truncateBlocktimes  = `TRUNCATE tracelistener.blocktime`
)

var migrations = []string{
	createTraceListenerDatabase,
	createDenomTracesTable,
	createChannelsTable,
	createConnectionsTable,
	createClientsTable,
	createBlockTimeTable,
}

var truncating = []string{
	truncateDenomTraces,
	truncateChannels,
	truncateConnections,
	truncateClients,
	truncateBlocktimes,
}

type TracelistenerData struct {
	Denoms      []DenomTrace
	Channels    []Channel
	Connections []Connection
	Clients     []Client
	BlockTimes  []BlockTime
}

type DenomTrace struct {
	Path      string
	BaseDenom string
	Hash      string
	ChainName string
}

type Channel struct {
	ChannelID        string
	CounterChannelID string
	Port             string
	State            int
	Hops             []string
	ChainName        string
}

type Connection struct {
	ChainName           string
	ConnectionID        string
	ClientID            string
	State               string
	CounterConnectionID string
	CounterClientID     string
}

type Client struct {
	SourceChainName string
	DestChainID     string
	ClientID        string
	LatestHeight    string
	TrustingPeriod  string
}

type BlockTime struct {
	ChainName string
	Time      time.Time
}

// TestingCtx A struct to hold context for child tests
type TestingCtx struct {
	Router *router.Router
	Cfg    *config.Config
	CnsDB  *cnsDb.Instance
}

// Setup Set up HTTP server, CDB and Redis in new ports.
// K8s clients are mocked.
func Setup() *TestingCtx {

	c := &config.Config{
		DatabaseConnectionURL: "FILLME",
		ListenAddr:            "FILLME",
		RedisAddr:             "FILLME",
		KubernetesNamespace:   k8sNsInTest,
		Debug:                 true,
	}

	l := logging.New(logging.LoggingConfig{
		LogPath: "",
		Debug:   c.Debug,
	})

	l.Infow("api-server", "version", "test")

	// --- CDB ---
	cdbTestServer, err := testserver.NewTestServer()
	checkNoError(err, l)

	checkNoError(cdbTestServer.WaitForInit(), l)

	c.DatabaseConnectionURL = cdbTestServer.PGURL().String()
	checkNotNil(c.DatabaseConnectionURL, "CDB conn. string", l)

	// FIXME: Do NOT initialize and migrate the DB using the CNS server's connection method
	// A big no-no here, using one service's internals inside the other
	// But no other way, since one service writes and the other reads, sharing the DB schemas
	cns, err := cnsDb.New(c.DatabaseConnectionURL)
	checkNoError(err, l)

	dbi, err := apiDb.Init(c)
	checkNoError(err, l)

	// --- Redis ---
	miniRedis, err := miniredis.Run()
	checkNoError(err, l)
	c.RedisAddr = miniRedis.Addr()
	s, err := store.NewClient(c.RedisAddr)
	checkNoError(err, l)

	// --- K8s ---
	kube := mocks.Client{}
	informer := mocks.GenericInformer{}

	r := router.New(
		dbi,
		l,
		s,
		&kube,
		c.KubernetesNamespace,
		&informer,
		c.Debug,
	)

	// --- HTTP server ---
	port, err := getFreePort()
	checkNoError(err, l)
	c.ListenAddr = "127.0.0.1:" + port

	ch := make(chan struct{})
	go func() {
		close(ch)
		err := r.Serve(c.ListenAddr)
		checkNoError(err, l)
	}()
	<-ch // Wait for the goroutine to start. Still hack!!

	return &TestingCtx{
		Cfg:    c,
		Router: r,
		CnsDB:  cns,
	}
}

// Creates tracelistner database and required tables only if they dont exist
func RunTraceListnerMigrations(ctx *TestingCtx, t *testing.T) {
	for _, m := range migrations {
		_, err := ctx.CnsDB.Instance.DB.Exec(m)
		require.NoError(t, err)
	}
}

// Empties all tracelistner tables
func TruncateTracelistener(ctx *TestingCtx, t *testing.T) {
	for _, m := range truncating {
		_, err := ctx.CnsDB.Instance.DB.Exec(m)
		require.NoError(t, err)
	}
}

// runs the given qurey with args and checks affected rows != 0
func insertRow(ctx *TestingCtx, t *testing.T, query string, args ...interface{}) {

	res, err := ctx.CnsDB.Instance.DB.Exec(query, args...)
	require.NoError(t, err)

	rows, _ := res.RowsAffected()

	require.NotEqual(t, 0, rows)
}

//	inserts data from given struct into respective tracelistener tables
func InsertTraceListnerData(ctx *TestingCtx, t *testing.T, data TracelistenerData) {
	for _, d := range data.Denoms {
		insertRow(ctx, t, insertDenomTrace, d.Path, d.BaseDenom, d.Hash, d.ChainName)
	}

	for _, d := range data.Channels {
		insertRow(ctx, t, insertChannel, d.ChannelID, d.CounterChannelID, d.Port, d.State, d.Hops, d.ChainName)
	}

	for _, d := range data.Connections {
		insertRow(ctx, t, insertConnection, d.ChainName, d.ConnectionID, d.ClientID, d.State, d.CounterConnectionID, d.CounterClientID)
	}

	for _, d := range data.Clients {
		insertRow(ctx, t, insertClient, d.SourceChainName, d.DestChainID, d.ClientID, d.LatestHeight, d.TrustingPeriod)
	}

	for _, d := range data.BlockTimes {
		insertRow(ctx, t, insertBlocktime, d.ChainName, d.Time)
	}
}

// TruncateDB Empties the CNS DB of data.
// Only use in tests executed sequentially
func TruncateCNSDB(ctx *TestingCtx, t *testing.T) {
	// FIXME: Using DB util from another service
	_, err := ctx.CnsDB.Instance.DB.Exec("TRUNCATE cns.chains")
	require.NoError(t, err)
}

func getFreePort() (port string, err error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		return "", err
	}

	_, port, _ = net.SplitHostPort(ln.Addr().String())
	_ = ln.Close()

	return port, nil
}

func checkNoError(err error, logger *zap.SugaredLogger) {
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
}

func checkNotNil(obj interface{}, whatObj string, logger *zap.SugaredLogger) {
	if obj == nil {
		logger.Error(fmt.Printf("Value is nil: %s", whatObj))
		os.Exit(-1)
	}
}
