package liquidity

import "github.com/allinbits/emeris-utils/exported/sdktypes"

type SwapFeesResponse struct {
	Fees sdktypes.Coins `json:"fees"`
}
