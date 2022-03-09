package liquidity

import "github.com/emerishq/emeris-utils/exported/sdktypes"

type SwapFeesResponse struct {
	Fees sdktypes.Coins `json:"fees"`
}
