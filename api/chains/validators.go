package chains

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/lib/ginutils"
	"github.com/emerishq/demeris-api-server/lib/keybase"
	"github.com/emerishq/demeris-api-server/lib/stringcache"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	avatarCacheDuration = 24 * time.Hour
	avatarCachePrefix   = "api-server/validator-avatars"
)

// GetValidators returns the list of validators.
// @Summary Gets list of validators of a specific chain.
// @Tags Chain
// @ID validators
// @Description Gets list of validators for a chain.
// @Description
// @Description These are the numerical value  correspondence of validator status.
// @Description 0: "BOND_STATUS_UNSPECIFIED"
// @Description	1: "BOND_STATUS_UNBONDED"
// @Description	2: "BOND_STATUS_UNBONDING"
// @Description	3: "BOND_STATUS_BONDED"
// @Produce json
// @Success 200 {object} ValidatorsResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /validators [get]
func GetValidators(d *deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := ginutils.GetValue[*zap.SugaredLogger](c, logging.LoggerKey)
		var res ValidatorsResponse

		chainName := c.Param("chain")
		validators, err := d.Database.GetValidators(chainName)
		if err != nil {
			e := apierrors.New(
				"validators",
				fmt.Sprintf("cannot retrieve validators"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve validators: %w", err),
				"chain",
				chainName,
			)
			_ = c.Error(e)

			return
		}

		adaptValidators := make([]*Validator, 0, len(validators))
		avatarCache := stringcache.NewStringCache(
			logger,
			stringcache.NewStoreBackend(d.Store),
			avatarCacheDuration,
			avatarCachePrefix,
			stringcache.HandlerFunc(fetchKeybaseAvatar),
		)
		for _, v := range validators {
			adapted, err := adaptValidator(c.Request.Context(), avatarCache, v)
			if err != nil {
				logger.Warnw(
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
}

func adaptValidator(ctx context.Context, cache *stringcache.StringCache, r tracelistener.ValidatorRow) (*Validator, error) {
	var v = &Validator{ValidatorRow: r}
	var err error

	if len(r.Identity) > 0 {
		v.Avatar, err = cache.Get(ctx, r.Identity, true)
	}

	return v, err
}

func fetchKeybaseAvatar(ctx context.Context, key string) (string, error) {
	avatar, err := keybase.GetPictureByKeySuffix(ctx, key)
	if err != nil {
		err = fmt.Errorf("keybase api: %w", err)
		return "", err
	}

	return avatar, err
}
