package usecase_test

import (
	"errors"
	"testing"

	"github.com/emerishq/demeris-api-server/usecase"
	"github.com/golang/mock/gomock"
)

type mocks struct {
	t                 *testing.T
	db                *MockDB
	sdkServiceClients *MockSDKServiceClients
	sdkServiceClient  *MockSDKServiceClient
}

func newApp(t *testing.T, setup func(mocks)) *usecase.App {
	ctrl := gomock.NewController(t)
	m := mocks{
		t:                 t,
		db:                NewMockDB(ctrl),
		sdkServiceClients: NewMockSDKServiceClients(ctrl),
		sdkServiceClient:  NewMockSDKServiceClient(ctrl),
	}

	// Pre-setup expectations on sdkServiceClients
	m.sdkServiceClients.EXPECT().GetSDKServiceClient("42").
		Return(m.sdkServiceClient, nil).AnyTimes()
	m.sdkServiceClients.EXPECT().GetSDKServiceClient("44").
		Return(m.sdkServiceClient, nil).AnyTimes()
	m.sdkServiceClients.EXPECT().GetSDKServiceClient(gomock.Any()).
		Return(nil, errors.New("version not found")).AnyTimes()

	if setup != nil {
		setup(m)
	}
	return usecase.NewApp(m.db, m.sdkServiceClients)
}
