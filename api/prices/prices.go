package prices

import (
	"net/http"

	"github.com/emerishq/demeris-api-server/lib/ginutils"
	"github.com/emerishq/demeris-api-server/usecase"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PricesAPI struct {
	app App
}

func New(app App) *PricesAPI {
	return &PricesAPI{
		app: app,
	}
}

func Register(router *gin.Engine, app App) {
	api := New(app)

	router.Group("/prices").
		GET("/pools", api.GetPoolsPrices)
}

type GetPoolsPricesResponse struct {
	PoolResults usecase.PoolPricesResult `json:"pool_results"`
}

// GetPoolsPrices returns the list of available pools and their associated token's price.
// @Summary Get pools' tokens prices
// @Tags Price
// @ID get-pools-prices
// @Produce json
// @Success 200 {object} GetPoolsPricesResponse
// @Failure 500 {object} apierrors.UserFacingError
// @Router /prices/pools [get]
func (api *PricesAPI) GetPoolsPrices(c *gin.Context) {
	prices := api.app.PoolPrices(c.Request.Context())

	logger := ginutils.GetValue[*zap.SugaredLogger](c, logging.LoggerKey)
	for _, opt := range prices {
		if opt.Err != nil {
			logger.Errorw("fetching pool price", "err", opt.Err)
		}
	}

	c.JSON(http.StatusOK, GetPoolsPricesResponse{
		PoolResults: prices,
	})
}
