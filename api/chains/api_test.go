package chains_test

import (
	"testing"

	"github.com/emerishq/demeris-api-server/api/chains"
	"github.com/emerishq/demeris-api-server/mocks"
)

type mockeds struct {
	cacheBackend *mocks.CacheBackend
	app          *mocks.App
}

func newChainAPI(t *testing.T, setup func(mockeds)) *chains.ChainAPI {
	m := mockeds{
		cacheBackend: mocks.NewCacheBackend(t),
		app:          mocks.NewApp(t),
	}
	if setup != nil {
		setup(m)
	}
	return chains.New(m.cacheBackend, m.app)
}
