package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/block"
	"github.com/emerishq/demeris-api-server/api/cached"
	"github.com/emerishq/demeris-api-server/api/liquidity"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"k8s.io/client-go/informers"

	"github.com/emerishq/demeris-api-server/api/relayer"

	"github.com/emerishq/emeris-utils/logging"
	"github.com/emerishq/emeris-utils/validation"
	"github.com/gin-gonic/gin/binding"

	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/emerishq/demeris-api-server/api/chains"
	"github.com/emerishq/demeris-api-server/api/tx"
	"github.com/emerishq/demeris-api-server/api/verifieddenoms"

	"github.com/emerishq/demeris-api-server/api/account"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/emeris-utils/store"
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

	engine.Use(logging.CorrelationIDMiddleware(l))
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
	defer func() {
		if rval := recover(); rval != nil {
			// okay we panic-ed, log it through r's logger and write back internal server error
			err := apierrors.New(
				"fatal_error",
				fmt.Sprintf("internal server error"),
				http.StatusInternalServerError)

			logger := logging.AddCorrelationIDToLogger(c, r.l)
			logger.Errorw(
				"panic handler triggered while handling call",
				"endpoint", c.Request.RequestURI,
				"error", fmt.Sprint(rval),
			)

			userError := apierrors.NewUserFacingError(tryGetIntCorrelationID(c), err)
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				userError,
			)

			return
		}
	}()
	c.Next()
}

func (r *Router) decorateCtxWithDeps(c *gin.Context) {
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

	err := &apierrors.Error{}
	if !errors.As(l, &err) {
		panic(fmt.Sprintf("expected to receive error of type *apierrors.Errors, got %T with content: %v", l, l))
	}

	keysAndValues := append(err.LogKeysAndValues, "error", err)
	d := deps.GetDeps(c)
	d.Logger.Errorw(
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
