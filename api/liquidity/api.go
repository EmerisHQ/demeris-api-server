package liquidity

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, d *deps.Deps) {
	group := router.Group("/pool")

	group.GET("/:poolId/swapfees", getSwapFee(d))
}

// getSwapFee returns the swap fee of past 1 hour n.
// @Summary Gets swap fee by pool id.
// @Tags pool
// @ID swap fee
// @Description Gets swap fee of past one hour by pool id.
// @Param pool path string true "pool id"
// @Produce json
// @Success 200 {object} SwapFeesResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /pool/{poolID}/swapfees [get]
func getSwapFee(d *deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {

		poolId := c.Param("poolId")

		res, err := d.Store.GetSwapFees(poolId)
		if err != nil {
			e := apierrors.New(
				"swap fees",
				fmt.Sprintf("cannot get swap fees"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot get swap fees: %w", err),
				"poolId",
				poolId,
			)
			_ = c.Error(e)

			return
		}

		fees := sdktypes.Coins{}
		for _, f := range res {
			fees = append(fees, sdktypes.NewInt64Coin(f.Denom, f.Amount.Int64()))
		}

		c.JSON(http.StatusOK, SwapFeesResponse{Fees: fees})
	}
}
