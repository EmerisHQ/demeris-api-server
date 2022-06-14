package chains_test

import (
	"testing"

	"github.com/emerishq/demeris-api-server/api/chains"
	gomock "github.com/golang/mock/gomock"
)

type mocks struct {
	cacheBackend *MockCacheBackend
	app          *MockApp
}

func newChainAPI(t *testing.T, setup func(mocks)) *chains.ChainAPI {
	ctrl := gomock.NewController(t)
	m := mocks{
		cacheBackend: NewMockCacheBackend(ctrl),
		app:          NewMockApp(ctrl),
	}
	if setup != nil {
		setup(m)
	}
	return chains.New(m.cacheBackend, m.app)
}
