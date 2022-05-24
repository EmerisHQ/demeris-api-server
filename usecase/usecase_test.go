package usecase_test

import (
	"testing"

	"github.com/emerishq/demeris-api-server/mocks"
	"github.com/emerishq/demeris-api-server/usecase"
)

type mockeds struct {
	t         *testing.T
	sdkClient *mocks.SDKService
}

func newApp(t *testing.T, setup func(mockeds)) *usecase.App {
	m := mockeds{
		t:         t,
		sdkClient: mocks.NewSDKService(t),
	}
	if setup != nil {
		setup(m)
	}
	return usecase.NewApp(m.sdkClient)
}
