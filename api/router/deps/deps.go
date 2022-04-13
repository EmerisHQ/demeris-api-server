package deps

import (
	"k8s.io/client-go/informers"
	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/lib/ginutils"
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
	deps := ginutils.GetValue[*Deps](c, "deps")

	// override logger with the one from the gin.context
	logger, err := logging.GetLoggerFromContext(c)
	if err == nil {
		deps.Logger = logger
	} else {
		deps.Logger.Warnw("couldn't get logger from context, using fallback", "error", err)
	}

	return deps
}

// WriteError adds an error to the gin context and logs it.
func (d *Deps) WriteError(c *gin.Context, err *apierrors.Error, logMessage string, keyAndValues ...interface{}) {
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
