package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"

	"github.com/allinbits/demeris-api-server/api/config"
	"github.com/allinbits/demeris-api-server/api/database"
	"github.com/allinbits/demeris-api-server/api/router"
	"github.com/allinbits/demeris-api-server/utils/k8s"
	"github.com/allinbits/demeris-api-server/utils/logging"
	"github.com/allinbits/demeris-api-server/utils/store"
	gaia "github.com/cosmos/gaia/v5/app"
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
			err := http.ListenAndServe(":6060", nil)
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

	kubeClient, err := k8s.NewInCluster()
	if err != nil {
		l.Panicw("cannot initialize k8s", "error", err)
	}
	cdc, _ := gaia.MakeCodecs()

	l.Infow("setup relayers informer", "namespace", cfg.KubernetesNamespace)
	infConfig, err := rest.InClusterConfig()
	if err != nil {
		l.Panicw("k8s server panic", "error", err)
	}

	informer, err := k8s.GetInformer(infConfig, cfg.KubernetesNamespace, "relayers")
	if err != nil {
		l.Panicw("k8s server panic", "error", err)
	}

	go informer.Informer().Run(make(chan struct{}))

	r := router.New(
		dbi,
		l,
		s,
		kubeClient,
		cfg.KubernetesNamespace,
		cdc,
		informer,
		cfg.Debug,
	)

	if err := r.Serve(cfg.ListenAddr); err != nil {
		l.Panicw("http server panic", "error", err)
	}
}
