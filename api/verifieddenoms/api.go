package verifieddenoms

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/verified_denoms", GetVerifiedDenoms)
}

// GetVerifiedDenoms returns the list of verified denoms.
// @Summary Gets verified denoms
// @Tags Denoms
// @ID verified-denoms
// @Description gets verified denoms
// @Produce json
// @Success 200 {object} VerifiedDenomsResponse
// @Failure 500,403 {object} deps.Error
// @Router /verified_denoms [get]
func GetVerifiedDenoms(c *gin.Context) {
	var res VerifiedDenomsResponse

	d := deps.GetDeps(c)

	chains, err := d.Database.Chains()

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
