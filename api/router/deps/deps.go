package deps

import (
	"k8s.io/client-go/informers"
	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/ginutils"
	"github.com/emerishq/emeris-utils/store"
	"github.com/gin-gonic/gin"
)

// Deps represents a set of objects useful during the lifecycle of REST endpoints.
type Deps struct {
	Database         *database.Database
	Store            *store.Store
	K8S              *kube.Client
	RelayersInformer informers.GenericInformer
	KubeNamespace    string
}

func GetDeps(c *gin.Context) *Deps {
	deps := ginutils.GetValue[*Deps](c, "deps")
	return deps
}
