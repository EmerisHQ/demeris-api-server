package usecase_test

import (
	"errors"
	"testing"

	"github.com/emerishq/demeris-api-server/mocks"
	"github.com/emerishq/demeris-api-server/usecase"
	"github.com/stretchr/testify/mock"
)

type mockeds struct {
	t                 *testing.T
	sdkServiceClients *mocks.SDKServiceClients
	sdkService        *mocks.SDKService
}

func newApp(t *testing.T, setup func(mockeds)) usecase.IApp {
	m := mockeds{
		t:                 t,
		sdkServiceClients: mocks.NewSDKServiceClients(t),
		sdkService:        mocks.NewSDKService(t),
	}

	// Pre-setup expectations on sdkServiceClients
	m.sdkServiceClients.EXPECT().GetSDKServiceClient("42").Return(m.sdkService, nil).Maybe()
	m.sdkServiceClients.EXPECT().GetSDKServiceClient("44").Return(m.sdkService, nil).Maybe()
	m.sdkServiceClients.EXPECT().GetSDKServiceClient(mock.Anything).Return(nil, errors.New("version not found")).Maybe()

	if setup != nil {
		setup(m)
	}
	return usecase.NewApp(m.sdkServiceClients)
}
