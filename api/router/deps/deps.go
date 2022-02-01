package deps

import (
	"fmt"

	"k8s.io/client-go/informers"
	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/allinbits/demeris-api-server/api/apierror"
	"github.com/allinbits/demeris-api-server/api/database"
	"github.com/allinbits/emeris-utils/store"
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

	return deps
}

// WriteError logs and return client-facing errors
func (d *Deps) WriteError(c *gin.Context, err apierror.Error, logMessage string, keyAndValues ...interface{}) {
	_ = c.Error(err)
	d.LogError(logMessage, keyAndValues...)
}

// LogError logs errors internally without altering the http response
func (d *Deps) LogError(logMessage string, keyAndValues ...interface{}) {
	if keyAndValues != nil {
		d.Logger.Errorw(
			logMessage,
			keyAndValues...,
		)
	}
}
