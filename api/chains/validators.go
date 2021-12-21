package chains

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/gin-gonic/gin"
)

// GetValidators returns the list of validators.
// @Summary Gets list of validators of a specific chain.
// @Tags Chain
// @ID validators
// @Description Gets list of validators for a chain.
// @Produce json
// @Success 200 {object} validatorsResponse
// @Failure 500,403 {object} deps.Error
// @Router /validators [get]
func GetValidators(c *gin.Context) {
	var res validatorsResponse

	d := deps.GetDeps(c)
	chainName := c.Param("chain")

	if _, err := d.Database.Chain(chainName); err != nil {
		e := deps.NewError(
			"fee",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	validators, err := d.Database.GetValidators(chainName)

	if err != nil {
		e := deps.NewError(
			"validators",
			fmt.Errorf("cannot retrieve validators"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve validators",
			"id",
			e.ID,
			"error",
			err,
			"chain",
			chainName,
		)

		return
	}

	res.Validators = validators

	c.JSON(http.StatusOK, res)
}
