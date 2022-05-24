package router

import (
	"errors"
	"fmt"

	"github.com/emerishq/demeris-api-server/api/block"
	"github.com/emerishq/demeris-api-server/api/cached"
	"github.com/emerishq/demeris-api-server/api/liquidity"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/lib/stringcache"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/demeris-api-server/usecase"
	"k8s.io/client-go/informers"

	"github.com/emerishq/demeris-api-server/api/relayer"

	"github.com/emerishq/emeris-utils/logging"
	"github.com/emerishq/emeris-utils/sentryx"
	"github.com/emerishq/emeris-utils/validation"
	"github.com/gin-gonic/gin/binding"

	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/emerishq/demeris-api-server/api/chains"
	"github.com/emerishq/demeris-api-server/api/tx"
	"github.com/emerishq/demeris-api-server/api/verifieddenoms"

	"github.com/emerishq/demeris-api-server/api/account"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/emeris-utils/store"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	g  *gin.Engine
	DB *database.Database
	l  *zap.SugaredLogger
	s  *store.Store
}

func New(
	db *database.Database,
	l *zap.SugaredLogger,
	s *store.Store,
	kubeClient kube.Client,
	kubeNamespace string,
	genericInformer informers.GenericInformer,
	sdkServiceClients sdkservice.SDKServiceClients,
	app usecase.IApp,
	debug bool,
) *Router {
	gin.SetMode(gin.ReleaseMode)

	if debug {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	engine.Use(logging.AddLoggerMiddleware(l))
	r := &Router{
		g:  engine,
		DB: db,
		l:  l,
		s:  s,
	}

	r.metrics()

	validation.JSONFields(binding.Validator)

	if debug {
		engine.Use(logging.LogRequest(l.Desugar()))
	}

	engine.Use(ginzap.RecoveryWithZap(l.Desugar(), true))
	engine.Use(r.handleErrors)
	engine.Use(sentryx.GinMiddleware)
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false

	relayersInformer := relayer.NewInformer(genericInformer, kubeNamespace)

	registerRoutes(engine, r.DB, r.s, relayersInformer, sdkServiceClients, app)

	return r
}

func (r *Router) Serve(address string) error {
	return r.g.Run(address)
}

func (r *Router) handleErrors(c *gin.Context) {
	c.Next()

	l := c.Errors.Last()
	if l == nil {
		c.Next()
		return
	}

	err := &apierrors.Error{}
	if !errors.As(l, &err) {
		panic(fmt.Sprintf("expected to receive error of type *apierrors.Errors, got %T with content: %v", l, l))
	}

	keysAndValues := append(err.LogKeysAndValues, "error", err)

	logger, loggerErr := logging.GetLoggerFromContext(c)
	if loggerErr != nil {
		panic(fmt.Sprintf("gin context didn't contain logger: %s", loggerErr))
	}
	logger.Errorw(
		err.Error(),
		keysAndValues...,
	)

	id := tryGetIntCorrelationID(c)
	userError := apierrors.NewUserFacingError(id, err)
	c.JSON(err.StatusCode, userError)
}

func tryGetIntCorrelationID(c *gin.Context) string {
	id, ok := c.Request.Context().Value(logging.IntCorrelationIDName).(string)
	if !ok {
		return ""
	}
	return id
}

func registerRoutes(engine *gin.Engine, db *database.Database, s *store.Store,
	relayersInformer *relayer.Informer, sdkServiceClients sdkservice.SDKServiceClients,
	app usecase.IApp) {
	// @tag.name Account
	// @tag.description Account-querying endpoints
	account.Register(engine, db, s, sdkServiceClients)

	// @tag.name Denoms
	// @tag.description Denoms-related endpoints
	verifieddenoms.Register(engine, db)

	// @tag.name Chain
	// @tag.description Chain-related endpoints
	chains.Register(engine, db, stringcache.NewStoreBackend(s), sdkServiceClients, app)

	// @tag.name Transactions
	// @tag.description Transaction-related endpoints
	tx.Register(engine, db, s, sdkServiceClients)

	// @tag.name Relayer
	// @tag.description Relayer-related endpoints
	relayer.Register(engine, db, relayersInformer)

	// @tag.name Block
	// @tag.description Blocks-related endpoints
	block.Register(engine, db, s)

	// @tag.name liquidity
	// @tag.description pool-related endpoints
	liquidity.Register(engine, db, s)

	// @tag.name cached
	// @tag.description cached data endpoints
	cached.Register(engine, db, s)
}
