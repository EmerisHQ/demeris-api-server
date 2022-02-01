package liquidity

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-api-server/api/apierror"
	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/allinbits/emeris-utils/exported/sdktypes"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	group := router.Group("/pool")

	group.GET("/:poolId/swapfees", getSwapFee)
}

// getSwapFee returns the swap fee of past 1 hour n.
// @Summary Gets swap fee by pool id.
// @Tags pool
// @ID swap fee
// @Description Gets swap fee of past one hour by pool id.
// @Param pool path string true "pool id"
// @Produce json
// @Success 200 {object} SwapFeesResponse
// @Failure 500,403 {object} deps.Error
// @Router /pool/{poolID}/swapfees [get]
func getSwapFee(c *gin.Context) {

	d := deps.GetDeps(c)

	poolId := c.Param("poolId")

	res, err := d.Store.GetSwapFees(poolId)
	if err != nil {
		e := apierror.New(
			"swap fees",
			fmt.Errorf("cannot get swap fees"),
			http.StatusBadRequest,
		)

		apierror.WriteError(d.Logger, c, e,
			"cannot get swap fees",
			"id",
			e.ID,
			"poolId",
			poolId,
			"error",
			err,
		)

		return
	}

	fees := sdktypes.Coins{}
	for _, f := range res {
		fees = append(fees, sdktypes.NewInt64Coin(f.Denom, f.Amount.Int64()))
	}

	c.JSON(http.StatusOK, SwapFeesResponse{Fees: fees})
}
