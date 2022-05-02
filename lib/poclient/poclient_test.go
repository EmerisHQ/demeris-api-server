package poclient_test

import (
	"testing"

	"github.com/emerishq/demeris-api-server/lib/poclient"
	"github.com/stretchr/testify/require"
)

const poURL = "https://api.emeris.com/v1/oracle"

func TestGetPrice(t *testing.T) {
	tests := []struct {
		name      string
		poBaseURL string
		symbol    string
		success   bool
	}{
		{
			"invalid price oracle base url",
			"http://invalid.com",
			"symbol",
			false,
		},
		{
			"invalid symbol",
			poURL,
			"symbol",
			false,
		},
		{
			"valid token symbol",
			poURL,
			"ATOMUSDT",
			true,
		},
		{
			"valid fiat symbol",
			poURL,
			"USDEUR",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := poclient.NewPOClient(tt.poBaseURL)
			price, err := client.GetPrice(tt.symbol)
			if tt.success {
				require.NoError(t, err)
				require.NotEmpty(t, price)
				require.Positive(t, price.Price)
			} else {
				require.Error(t, err)
				require.Empty(t, price)
			}
		})
	}
}
