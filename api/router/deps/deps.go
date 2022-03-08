package deps

import (
	"fmt"

	"k8s.io/client-go/informers"
	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/emerishq/emeris-utils/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Deps represents a set of objects useful during the lifecycle of REST endpoints.
type Deps struct {
	Logger           *zap.SugaredLogger
	Database         *database.Database
	Store            *store.Store
	K8S              *kube.Client
	RelayersInformer informers.GenericInformer
	KubeNamespace    string
}

func GetDeps(c *gin.Context) *Deps {
	d, ok := c.Get("deps")
	if !ok {
		panic("deps not set in context")
	}

	deps, ok := d.(*Deps)
	if !ok {
		panic(fmt.Sprintf("deps not of the expected type, found %T", deps))
	}

	// override logger with the one from the gin.context
	logger, err := logging.GetLoggerFromContext(c)
	if err == nil {
		deps.Logger = logger
	} else {
		deps.Logger.Warnw("couldn't get logger from context, using fallback", "error", err)
	}

	return deps
}

// WriteError logs and return client-facing errors
func (d *Deps) WriteError(c *gin.Context, err Error, logMessage string, keyAndValues ...interface{}) {

	// setting error id
	value, ok := c.Request.Context().Value(logging.IntCorrelationIDName).(string)
	if !ok {
		panic("cant get value int_correlation_id")
	}
	err.ID = value
	_ = c.Error(err)

	if keyAndValues != nil {
		keyAndValues = append(keyAndValues, "error", err)
		d.Logger.Errorw(
			logMessage,
			keyAndValues...,
		)
	}
}

// LogError is used to log errors internally while returning 200 in the response
func (d *Deps) LogError(logMessage string, keyAndValues ...interface{}) {
	if keyAndValues != nil {
		d.Logger.Errorw(
			logMessage,
			keyAndValues...,
		)
	}
}
