package tx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/sdkservice"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// paths for packet_sequence and tx from the events,
// can be updated to more readable format after typed events are introduced in sdk
const (
	// packetSequencePath may need updates if ibc-go/transfer events change
	packetSequencePath = `tx_response.logs.#.events.#(type=="send_packet").attributes.#(key=="packet_sequence").value`
	// txPath may need updates if tendermint response changes
	txPath  = "result.txs.0.hash"
	timeout = 10 * time.Second
)

// getIBCSeqFromTx returns a list of sequence numbers gotten from data,
// from the sent IBC packet event.
// If no IBC sequence numbers are found, the resulting slice is empty.
func getIBCSeqFromTx(data []byte) []string {
	raw := gjson.GetBytes(data, packetSequencePath).Array()

	ret := make([]string, 0, len(raw))
	for _, r := range raw {
		ret = append(ret, r.String())
	}

	return ret
}

// GetDestTx returns tx hash on destination chain.
// @Summary Gets tx hash on destination chain.
// @Tags Tx
// @ID destTx
// @Description Gets tx hash on destination chain.
// @Param srcChain path string true "source chain name"
// @Param destChainName path string true "destination chain name"
// @Param txHash path string true "tx hash on src chain"
// @Produce json
// @Success 200 {object} DestTxResponse
// @Failure 500,403 {object} deps.Error
// @Router /tx/{srcChain}/{destChain}/{txHash} [get]
func GetDestTx(c *gin.Context) {
	d := deps.GetDeps(c)

	srcChain := c.Param("src-chain")
	destChain := c.Param("dest-chain")
	txHash := c.Param("tx-hash")

	srcChainInfo, err := d.Database.Chain(srcChain)
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve srcChainInfo with name %v", srcChain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve srcChainInfo",
			"id",
			e.ID,
			"name",
			srcChain,
			"error",
			err,
		)

		return
	}

	// validate destination srcChainInfo is present
	destChainInfo, err := d.Database.Chain(destChain)
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve srcChainInfo with name %v", destChain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve srcChainInfo",
			"id",
			e.ID,
			"name",
			destChain,
			"error",
			err,
		)

		return
	}

	client, err := sdkservice.Client(srcChainInfo.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with srcChainInfo name %v", srcChainInfo.CosmosSDKVersion, srcChainInfo.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve srcChainInfo's sdk-service",
			"id",
			e.ID,
			"name",
			srcChain,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.QueryTx(context.Background(), &sdkutilities.QueryTxPayload{
		ChainName: srcChainInfo.ChainName,
		Hash:      txHash,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve tx from sdk-service, %w", err),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve tx from sdk-service",
			"id",
			e.ID,
			"txHash",
			txHash,
			"src srcChainInfo name",
			srcChain,
			"error",
			err,
		)

		return
	}

	// This query always returns an array of sequence numbers.
	// Emeris-generated IBC transfers are always sent out alone, meaning that
	// there are no more than 1 IBC transfer per tx.
	// This code is ready to be adapted to support multiple IBC transfer/transaction, but
	// for now we just get the first seq number found and roll with it.
	r := getIBCSeqFromTx(sdkRes)
	if len(r) == 0 {
		e := deps.NewError(
			"chains",
			fmt.Errorf("provided transaction is not ibc transfer"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"provided transaction is not ibc transfer",
			"id",
			e.ID,
			"txHash",
			txHash,
			"src srcChainInfo name",
			srcChain,
			"error",
			err,
		)

		return
	}

	seqNum := r[0]
	url := fmt.Sprintf("http://%s:26657/tx_search?query=\"recv_packet.packet_sequence=%s\"", destChainInfo.ChainName, seqNum)

	httpClient := &http.Client{
		Timeout: timeout,
	}

	// we're validating inputs and hence gosec-G107 can be ignored
	resp, err := httpClient.Get(url) // nolint: gosec
	if err != nil || resp.StatusCode != http.StatusOK {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve tx with packet sequence %s on %s", seqNum, destChain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve destination tx",
			"id",
			e.ID,
			"txHash",
			txHash,
			"dest srcChainInfo name",
			destChain,
			"error",
			err,
			"status_code",
			resp.Status,
		)

		return
	}
	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve tx with packet sequence %s on %s", seqNum, destChain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve destination tx",
			"id",
			e.ID,
			"txHash",
			txHash,
			"dest srcChainInfo name",
			destChain,
			"error",
			err,
		)

		return
	}

	otherSideTxHash := gjson.GetBytes(bz, txPath)
	c.JSON(http.StatusOK, DestTxResponse{
		DestChain: destChain,
		TxHash:    otherSideTxHash.String(),
	})
}
