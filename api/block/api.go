package block

import (
	"fmt"
	"net/http"
	"strconv"

	// needed for swagger gen
	_ "encoding/json"

	"github.com/emerishq/emeris-utils/store"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, d *deps.Deps) {
	router.GET("/block_results", GetBlock(d))
}

// GetBlock returns a Tendermint block data at a given height.
// @Summary Returns block data at a given height.
// @Tags Block
// @ID get-block
// @Description returns block data at a given height
// @Produce json
// @Param height query string true "height to query for"
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /block_results [get]
func GetBlock(d *deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {

		h := c.Query("height")
		if h == "" {
			e := apierrors.New(
				"block",
				fmt.Sprintf("missing height"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query block, missing height"),
			)
			_ = c.Error(e)
			return
		}

		hh, err := strconv.ParseInt(h, 10, 64)
		if err != nil {
			e := apierrors.New(
				"block",
				fmt.Sprintf("malformed height"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query block, malformed height: %w", err),
				"height_string",
				h,
			)
			_ = c.Error(e)
			return
		}

		bs := store.NewBlocks(d.Store)

		bd, err := bs.Block(hh)
		if err != nil {
			e := apierrors.New(
				"block",
				fmt.Sprintf("cannot get block at height %v", hh),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query block from redis: %w", err),
				"height",
				hh,
			)
			_ = c.Error(e)
			return
		}

		c.Data(http.StatusOK, "application/json", bd)
	}
}
