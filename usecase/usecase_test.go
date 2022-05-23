package usecase_test

import (
	"testing"

	"github.com/emerishq/demeris-api-server/usecase"
	gomock "github.com/golang/mock/gomock"
)

type mocks struct {
	t         *testing.T
	sdkClient *MockSDKClient
}

func newApp(t *testing.T, setup func(mocks)) *usecase.App {
	ctrl := gomock.NewController(t)
	m := mocks{
		t:         t,
		sdkClient: NewMockSDKClient(ctrl),
	}
	if setup != nil {
		setup(m)
	}
	return usecase.NewApp(m.sdkClient)
}
