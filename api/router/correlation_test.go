package router_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/emerishq/demeris-api-server/api/config"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/api/router"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/emerishq/emeris-utils/store"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestCorrelationIDMiddleWare(t *testing.T) {
	t.Parallel()
	r, cfg, observedLogs, tDown := setup(t)
	defer tDown()
	require.NotNil(t, r)

	go r.Serve(cfg.ListenAddr)
	time.Sleep(2 * time.Second)

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s%s", cfg.ListenAddr, "/chain/foo"), nil)
	require.NoError(t, err)

	id, err := uuid.NewV4()
	require.NoError(t, err)

	req.Header.Set("X-Correlation-id", fmt.Sprintf("%x", id))

	_, err = client.Do(req)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		count := 0
		for _, info := range observedLogs.All() {
			if info.ContextMap()[string(logging.IntCorrelationIDName)] != nil {
				count++
			}
			if info.ContextMap()[string(logging.CorrelationIDName)] == fmt.Sprintf("%x", id) {
				count++
			}
		}
		return count == 4
	}, 5*time.Second, 1*time.Second)
}

func setup(t *testing.T) (router.Router, config.Config, *observer.ObservedLogs, func()) {
	tServer, err := testserver.NewTestServer()
	require.NoError(t, err)

	require.NoError(t, tServer.WaitForInit())

	connStr := tServer.PGURL().String()
	require.NotNil(t, connStr)

	cfg := &config.Config{
		DatabaseConnectionURL: connStr,
		ListenAddr:            "127.0.0.1:9090",
		RedisAddr:             "127.0.0.1:6379",
		KubernetesNamespace:   "emeris",
		Debug:                 true,
	}

	db, err := database.Init(cfg)
	require.NoError(t, err)

	s, err := store.NewClient(cfg.RedisAddr)
	require.NoError(t, err)

	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	observedLogger := zap.New(observedZapCore)

	clients, err := sdkservice.InitializeClients()
	require.NoError(t, err)

	return *router.New(db, observedLogger.Sugar(), s, nil, "", nil, clients, nil, cfg.Debug), *cfg, observedLogs, func() { tServer.Stop() }
}
