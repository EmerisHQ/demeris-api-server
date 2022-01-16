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

// GetDestTx returns the transaction status n.
// @Summary Gets ticket by id.
// @Tags Chain
// @ID txTicket
// @Description Gets transaction status by ticket id.
// @Param ticketId path string true "ticket id"
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} store.Ticket
// @Failure 500,403 {object} deps.Error
// @Router /tx/ticket/{chainName}/{ticketId} [get]
func GetDestTx(c *gin.Context) {
	d := deps.GetDeps(c)

	srcChain := c.Param("srcChain")
	destChain := c.Param("destChain")
	txHash := c.Param("txHash")

	url := fmt.Sprintf("http://%s:26657/tx?hash=%s&prove=%t", srcChain, "0x"+txHash, false)
	resp, err := http.Get(url)
	if err != nil {
		d.LogError("http get", err)
		return
	}

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
	}
	r := gjson.GetBytes(bz, "result.tx_result.events.3.attributes.3.value")
	d.Logger.Debugw("thi si tx", "tx", r)

	seq, err := base64.StdEncoding.DecodeString(r.String())

	url = fmt.Sprintf("http://%s:26657/tx_search?query=\"recv_packet.packet_sequence=%s\"", destChain, string(seq))
	d.Logger.Debugw("this log", "url", url)
	resp, err = http.Get(url)
	if err != nil {
		d.LogError("http get", err)
		return
	}
	bz, err = io.ReadAll(resp.Body)
	if err != nil {
	}

	r = gjson.GetBytes(bz, "result.txs.0.hash")
	d.Logger.Debugw("thi si tx", "tx", r)

	c.JSON(http.StatusOK, r.String())
}
