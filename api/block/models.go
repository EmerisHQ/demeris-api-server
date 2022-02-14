package block

import (
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// nolint :deadcode being used for swagger generation
type BlockHeightResp struct {
	JSONRPC string                `json:"jsonrpc"`
	ID      string                `json:"id"`
	Result  coretypes.ResultBlock `json:"result"`
}
