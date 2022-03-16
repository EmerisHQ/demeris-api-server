package sdkservice_test

import (
	"fmt"
	"testing"

	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/stretchr/testify/require"
)

func TestSdkServiceURL(t *testing.T) {
	testCases := []struct {
		sdkVersion string
		sdkService string
	}{
		{
			"45",
			"sdk-service-v44:9090",
		},
		{
			"44",
			"sdk-service-v44:9090",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.sdkVersion, func(t *testing.T) {
			actual := sdkservice.SdkServiceURL(tt.sdkVersion)
			require.Equal(t, tt.sdkService, actual)
		})
	}

}
