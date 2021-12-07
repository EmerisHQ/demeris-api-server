package tx

import (
	"context"
	"fmt"

	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
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

	grpcConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", srcChain, 9090), // Or your gRPC server address.
		grpc.WithInsecure(),                  // The SDK doesn't support any transport security mechanism.
	)

	if err != nil {
		return
	}

	defer grpcConn.Close()

	txClient := tx.NewServiceClient(grpcConn)
	grpcRes, err := txClient.GetTx(context.Background(), &tx.GetTxRequest{Hash: txHash})
	if err != nil {
		return
	}
	tx := grpcRes.GetTx().GetAuthInfo().SignerInfos

}
