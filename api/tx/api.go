package tx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.POST("/tx/:chain", Tx)
	router.GET("/tx/:src-chain/:dest-chain/:tx-hash", GetDestTx)
	router.POST("/tx/:chain/simulate", GetTxFeeEstimate)
	router.GET("/tx/ticket/:chain/:ticket", GetTicket)
}

// Tx relays a transaction to an internal node for the specified chain.
// @Summary Relays a transaction to the relevant chain.
// @Tags Tx
// @ID tx
// @Description Relays a transaction to the relevant chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} TxResponse
// @Failure 500,400 {object} apierrors.UserFacingError
// @Router /tx/{chainName} [post]
func Tx(c *gin.Context) {
	// var tx typestx.Tx
	var txRequest TxRequest

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	err := c.BindJSON(&txRequest)

	if err != nil {
		e := apierrors.New("tx", fmt.Sprintf("failed to parse JSON"), http.StatusBadRequest).WithLogContext(
			fmt.Errorf("Failed to parse JSON: %w", err),
		)
		_ = c.Error(e)

		return
	}

	chain, err := d.Database.Chain(chainName)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chain: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusInternalServerError,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chain's sdk-service: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	txhash, err := relayTx(client, d, txRequest.TxBytes, chainName, txRequest.Owner)

	if err != nil {
		e := apierrors.New("tx", fmt.Sprintf("relaying tx failed, %v", err), http.StatusBadRequest).WithLogContext(
			fmt.Errorf("relaying tx failed: %w", err),
		)
		_ = c.Error(e)

		return
	}

	c.JSON(http.StatusOK, TxResponse{
		Ticket: txhash,
	})
}

// relayTx relays the tx to the specifc endpoint
// relayTx will also perform the ticketing mechanism
// Always expect broadcast mode to be `async`
func relayTx(services sdkutilities.Client, d *deps.Deps, txBytes []byte, chainName string, owner string) (string, error) {
	res, err := services.BroadcastTx(context.Background(), &sdkutilities.BroadcastTxPayload{
		ChainName: chainName,
		TxBytes:   txBytes,
	})

	if err != nil {
		return "", err
	}

	err = d.Store.CreateTicket(chainName, res.Hash, owner)

	if err != nil {
		return res.Hash, err
	}

	return res.Hash, nil
}

// GetTicket returns the transaction status n.
// @Summary Gets ticket by id.
// @Tags Chain
// @ID txTicket
// @Description Gets transaction status by ticket id.
// @Param ticketId path string true "ticket id"
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} store.Ticket
// @Failure 400 {object} apierrors.UserFacingError
// @Router /tx/ticket/{chainName}/{ticketId} [get]
func GetTicket(c *gin.Context) {

	d := deps.GetDeps(c)

	chainName := c.Param("chain")
	ticketId := c.Param("ticket")

	ticket, err := d.Store.Get(fmt.Sprintf("%s/%s", chainName, ticketId))

	if err != nil {
		e := apierrors.New(
			"tx",
			fmt.Sprintf("cannot retrieve ticket with id %v", ticketId),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve ticket: %w", err),
			"name",
			ticketId,
		)
		_ = c.Error(e)

		return
	}

	c.JSON(http.StatusOK, ticket)
}

// GetTxFeeEstimate returns the estimated gas and fee price for specified chain.
// @Summary estimates the gas and fees fot transaction.
// @Tags Tx
// @ID txFees
// @Description estimate transaction fees for the relevant chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} TxFeeEstimateRes
// @Failure 500,400 {object} apierrors.UserFacingError
// @Router /tx/fees/{chainName} [post]
func GetTxFeeEstimate(c *gin.Context) {
	var txRequest TxFeeEstimateReq

	d := deps.GetDeps(c)
	chainName := c.Param("chain")

	err := c.BindJSON(&txRequest)
	if err != nil {
		e := apierrors.New("tx", fmt.Sprintf("failed to parse JSON"), http.StatusBadRequest).WithLogContext(
			fmt.Errorf("Failed to parse JSON: %w", err),
		)
		_ = c.Error(e)

		return
	}

	chain, err := d.Database.Chain(chainName)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chain: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusInternalServerError,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chain's sdk-service: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	sdkRes, err := client.EstimateFees(context.Background(), &sdkutilities.EstimateFeesPayload{
		ChainName: chainName,
		TxBytes:   txRequest.TxBytes,
	})

	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot estimate fees from sdk-service"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot estimate fees from sdk-service: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	coins := sdktypes.Coins{}
	for _, c := range sdkRes.Fees {
		amt, _ := sdktypes.NewIntFromString(c.Amount)
		coins = append(coins, sdktypes.Coin{
			Denom:  c.Denom,
			Amount: amt,
		})
	}

	c.JSON(http.StatusOK, TxFeeEstimateRes{
		GasWanted: sdkRes.GasWanted,
		GasUsed:   sdkRes.GasUsed,
		Fees:      coins,
	})
}
