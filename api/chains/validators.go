package chains

import (
	"context"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/allinbits/demeris-api-server/lib/keybase"
	"github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/gin-gonic/gin"
)

// GetValidators returns the list of validators.
// @Summary Gets list of validators of a specific chain.
// @Tags Chain
// @ID validators
// @Description Gets list of validators for a chain.
// @Produce json
// @Success 200 {object} ValidatorsResponse
// @Failure 500,403 {object} deps.Error
// @Router /validators [get]
func GetValidators(c *gin.Context) {
	var res ValidatorsResponse

	d := deps.GetDeps(c)
	chainName := c.Param("chain")

	if exists, err := d.Database.ChainExists(chainName); err != nil || !exists {
		e := deps.NewError(
			"primarychannel",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		if err == nil {
			err = fmt.Errorf("%s chain doesnt exists", chainName)
		}

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

	adaptValidators := make([]*Validator, 0, len(validators))
	for _, v := range validators {
		adapted, err := adaptValidator(c.Request.Context(), v)
		if err != nil {
			d.Logger.Warnw(
				"cannot get avatar for validator",
				"validatorIdentity", v.Identity,
				"error", err,
			)
		}

		adaptValidators = append(adaptValidators, adapted)
	}

	res.Validators = adaptValidators

	c.JSON(http.StatusOK, res)
}

func adaptValidator(c context.Context, r tracelistener.ValidatorRow) (*Validator, error) {
	var v = &Validator{ValidatorRow: r}
	var err error

	if len(r.Identity) > 0 {
		v.Avatar, err = getAvatar(c, r.Identity)
	}

	return v, err
}

func getAvatar(c context.Context, keySuffix string) (string, error) {
	// TODO: introduce caching

	avatar, err := keybase.GetPictureByKeySuffix(c, keySuffix)
	if err != nil {
		// an error fetching avatar shouldn't block execution, we'll just log it
		err = fmt.Errorf("keybase api: %w", err)
	}

	return avatar, err
}
