package tx

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

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

	srcChain := c.Param("srcChain")
	destChain := c.Param("destChain")
	txHash := c.Param("txHash")

	url := fmt.Sprintf("http://%s:26657/tx?hash=%s&prove=%t", srcChain, "0x"+txHash, false)

	bz, err := GetUrlRes(url)
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve tx info of %s on %s", txHash, srcChain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve tx info",
			"id",
			e.ID,
			"txHash",
			txHash,
			"src chain name",
			srcChain,
			"error",
			err,
		)

		return
	}

	r := gjson.GetBytes(bz, "result.tx_result.events.3.attributes.3.value")
	seq, err := base64.StdEncoding.DecodeString(r.String())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve packet sequenceof %s on %s", txHash, srcChain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve sequence",
			"id",
			e.ID,
			"txHash",
			txHash,
			"src chain name",
			srcChain,
			"error",
			err,
		)

		return
	}

	url = fmt.Sprintf("http://%s:26657/tx_search?query=\"recv_packet.packet_sequence=%s\"", destChain, string(seq))

	bz, err = GetUrlRes(url)
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve tx with packet sequence %d on %s", seq, destChain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve destination tx",
			"id",
			e.ID,
			"txHash",
			txHash,
			"dest chain name",
			destChain,
			"error",
			err,
		)

		return
	}

	r = gjson.GetBytes(bz, "result.txs.0.hash")
	c.JSON(http.StatusOK, DestTxResponse{
		DestChain: destChain,
		TxHash:    r.String(),
	})
}

func GetUrlRes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bz, nil
}
