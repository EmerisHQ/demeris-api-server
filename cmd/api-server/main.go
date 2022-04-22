package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"

	"github.com/emerishq/demeris-api-server/api/config"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/api/router"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/emeris-utils/k8s"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/emerishq/emeris-utils/store"
	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"k8s.io/client-go/rest"
)

var Version = "not specified"

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: cfg.Debug,
	})

	if cfg.Debug {
		runtime.SetCPUProfileRate(500)

		go func() {
			http.HandleFunc("/freemem", func(_ http.ResponseWriter, _ *http.Request) {
				runtime.GC()
				debug.FreeOSMemory()
			})

			l.Debugw("starting profiling server", "port", "6060")
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				l.Panicw("cannot run profiling server", "error", err)
			}
		}()
	}

	l.Infow("api-server", "version", Version)

	dbi, err := database.Init(cfg)
	if err != nil {
		l.Panicw("cannot initialize database", "error", err)
	}

	s, err := store.NewClient(cfg.RedisAddr)
	if err != nil {
		l.Panicw("unable to start redis client", "error", err)
	}

	var k8sCfg *rest.Config
	if cfg.KubernetesConfigMode == "kubectl" {
		k8sCfg, err = k8s.KubectlConfig()
		if err != nil {
			l.Panicw("cannot get kubernetes config using kubectl", "error", err)
		}
	} else {
		k8sCfg, err = k8s.InClusterConfig()
		if err != nil {
			l.Panicw("cannot get kubernetes config. Are you running this service inside a pod?", "error", err)
		}
	}

	kubeClient, err := k8s.NewClient(k8sCfg)
	if err != nil {
		l.Panicw("cannot initialize kubernetes client", "error", err)
	}

	l.Infow("setup relayers informer", "namespace", cfg.KubernetesNamespace)
	informer, err := k8s.GetInformer(k8sCfg, cfg.KubernetesNamespace, "relayers")
	if err != nil {
		l.Panicw("k8s server panic", "error", err)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		SampleRate:       cfg.SentrySampleRate,
		TracesSampleRate: cfg.SentryTracesSampleRate,
		AttachStacktrace: true,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	go informer.Informer().Run(make(chan struct{}))

	sdkServiceClients, err := sdkservice.InitializeClients()
	if err != nil {
		l.Panicw("cannot initialize sdk-service clients", "error", err)
	}

	r := router.New(
		dbi,
		l,
		s,
		kubeClient,
		cfg.KubernetesNamespace,
		informer,
		sdkServiceClients,
		cfg.Debug,
	)

	if err := r.Serve(cfg.ListenAddr); err != nil {
		l.Panicw("http server panic", "error", err)
	}
}
