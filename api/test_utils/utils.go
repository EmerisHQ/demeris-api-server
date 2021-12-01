package test_utils

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/allinbits/emeris-utils/store"
	"github.com/stretchr/testify/require"

	"github.com/allinbits/demeris-api-server/api/config"
	apiDb "github.com/allinbits/demeris-api-server/api/database"
	cnsDb "github.com/allinbits/emeris-cns-server/cns/database"

	"github.com/alicebob/miniredis/v2"
	"github.com/allinbits/demeris-api-server/api/router"
	"github.com/allinbits/demeris-api-server/mocks"
	"github.com/allinbits/emeris-utils/logging"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"go.uber.org/zap"
)

const k8sNsInTest = "emeris"

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

// TruncateDB Empties the DB of data.
// Only use in tests executed sequentially
func TruncateDB(ctx *TestingCtx, t *testing.T) {
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
