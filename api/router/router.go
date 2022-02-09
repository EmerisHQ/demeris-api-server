package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-api-server/api/block"
	"github.com/allinbits/demeris-api-server/api/cached"
	"github.com/allinbits/demeris-api-server/api/liquidity"
	"k8s.io/client-go/informers"

	"github.com/allinbits/demeris-api-server/api/relayer"

	"github.com/allinbits/emeris-utils/logging"
	"github.com/allinbits/emeris-utils/validation"
	"github.com/gin-gonic/gin/binding"

	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/allinbits/demeris-api-server/api/chains"
	"github.com/allinbits/demeris-api-server/api/tx"
	"github.com/allinbits/demeris-api-server/api/verifieddenoms"

	"github.com/allinbits/demeris-api-server/api/account"
	"github.com/allinbits/demeris-api-server/api/database"
	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/allinbits/emeris-utils/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	g                *gin.Engine
	DB               *database.Database
	l                *zap.SugaredLogger
	s                *store.Store
	k8s              kube.Client
	k8sNamespace     string
	relayersInformer informers.GenericInformer
}

func New(
	db *database.Database,
	l *zap.SugaredLogger,
	s *store.Store,
	kubeClient kube.Client,
	kubeNamespace string,
	relayersInformer informers.GenericInformer,
	debug bool,
) *Router {
	gin.SetMode(gin.ReleaseMode)

	if debug {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	engine.Use(CorrelationIDMiddleware(l))

	r := &Router{
		g:                engine,
		DB:               db,
		l:                l,
		s:                s,
		k8s:              kubeClient,
		k8sNamespace:     kubeNamespace,
		relayersInformer: relayersInformer,
	}

	r.metrics()

	validation.JSONFields(binding.Validator)

	if debug {
		engine.Use(logging.LogRequest(l.Desugar()))
	}
	engine.Use(r.catchPanicsFunc)
	engine.Use(r.decorateCtxWithDeps)
	engine.Use(r.handleErrors)
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false

	registerRoutes(engine)

	return r
}

func (r *Router) Serve(address string) error {
	return r.g.Run(address)
}

func (r *Router) catchPanicsFunc(c *gin.Context) {
	l := GetLoggerFromContext(c)
	defer func() {
		if rval := recover(); rval != nil {
			// okay we panic-ed, log it through r's logger and write back internal server error
			err := deps.NewError(
				"fatal_error",
				errors.New("internal server error"),
				http.StatusInternalServerError)

			l.Errorw(
				"panic handler triggered while handling call",
				"endpoint", c.Request.RequestURI,
				"error", fmt.Sprint(rval),
				"error_id", err.ID,
			)

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				err,
			)

			return
		}
	}()
	c.Next()
}

func (r *Router) decorateCtxWithDeps(c *gin.Context) {
	r.l = GetLoggerFromContext(c)

	c.Set("deps", &deps.Deps{
		Logger:           r.l,
		Database:         r.DB,
		Store:            r.s,
		KubeNamespace:    r.k8sNamespace,
		K8S:              &r.k8s,
		RelayersInformer: r.relayersInformer,
	})
}

func (r *Router) handleErrors(c *gin.Context) {
	c.Next()

	l := c.Errors.Last()
	if l == nil {
		c.Next()
		return
	}

	rerr := deps.Error{}
	if !errors.As(l, &rerr) {
		panic(l)
	}

	c.JSON(rerr.StatusCode, rerr)
}

func registerRoutes(engine *gin.Engine) {
	// @tag.name Account
	// @tag.description Account-querying endpoints
	account.Register(engine)

	// @tag.name Denoms
	// @tag.description Denoms-related endpoints
	verifieddenoms.Register(engine)

	// @tag.name Chain
	// @tag.description Chain-related endpoints
	chains.Register(engine)

	// @tag.name Transactions
	// @tag.description Transaction-related endpoints
	tx.Register(engine)

	// @tag.name Relayer
	// @tag.description Relayer-related endpoints
	relayer.Register(engine)

	// @tag.name Block
	// @tag.description Blocks-related endpoints
	block.Register(engine)

	// @tag.name liquidity
	// @tag.description pool-related endpoints
	liquidity.Register(engine)

	// @tag.name cached
	// @tag.description cached data endpoints
	cached.Register(engine)

}
