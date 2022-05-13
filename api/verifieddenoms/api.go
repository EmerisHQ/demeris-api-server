package verifieddenoms

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, db *database.Database) {
	router.GET("/verified_denoms", GetVerifiedDenoms(db))
}

// GetVerifiedDenoms returns the list of verified denoms.
// @Summary Gets verified denoms
// @Tags Denoms
// @ID verified-denoms
// @Description gets verified denoms
// @Produce json
// @Success 200 {object} VerifiedDenomsResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /verified_denoms [get]
func GetVerifiedDenoms(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var res VerifiedDenomsResponse

		chains, err := db.Chains(ctx)

		if err != nil {
			e := apierrors.New(
				"verified_denoms",
				fmt.Sprintf("cannot retrieve chains"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve chains: %w", err),
			)
			_ = c.Error(e)

			return
		}

		for _, cc := range chains {
			for _, vd := range cc.VerifiedTokens() {
				res.VerifiedDenoms = append(res.VerifiedDenoms, VerifiedDenom{
					Denom:     vd,
					ChainName: cc.ChainName,
				})
			}
		}

		c.JSON(http.StatusOK, res)
	}
}
