package deps

import (
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/emeris-utils/store"
)

// Deps represents a set of objects useful during the lifecycle of REST endpoints.
type Deps struct {
	Database *database.Database
	Store    *store.Store
}
