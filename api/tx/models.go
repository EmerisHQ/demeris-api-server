package tx

import (
	"github.com/allinbits/demeris-backend-models/cns"
	"github.com/allinbits/emeris-utils/exported/sdktypes"
)

type TxRequest struct {
	Owner   string `json:"owner"`
	TxBytes []byte `json:"tx_bytes"`
}

type TxMeta struct {
	RelayOnly      bool
	TxType         string
	Signer         string
	SignerSequence string
	FeePayer       string
	Valid          bool
	Chain          cns.Chain
}

type TxResponse struct {
	Ticket string `json:"ticket"`
}

type TxFeeEstimateReq struct {
	TxBytes []byte `json:"tx_bytes"`
}

type TxFeeEstimateRes struct {
	GasWanted uint64
	GasUsed   uint64
	Fees      []sdktypes.Coin
}

type DestTxResponse struct {
	DestChain string `json:"dest_chain"`
	TxHash    string `json:"tx_hash"`
}
