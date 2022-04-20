package cached

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/emeris-utils/store"
	"github.com/gin-gonic/gin"
	_ "github.com/gravity-devs/liquidity/x/liquidity/types"
)

func Register(router *gin.Engine, db *database.Database, s *store.Store) {
	group := router.Group("/cached/cosmos")

	group.GET("/liquidity/v1beta1/pools", getPools(db, s))
	group.GET("/liquidity/v1beta1/params", getParams(db, s))
	group.GET("/bank/v1beta1/supply", getSupply(db, s))
	group.GET("/node_info", getNodeInfo(db, s))
}

// getPools returns the of all pools.
// @Summary Gets pools info.
// @Tags pools
// @ID pools
// @Description Gets info of all pools.`10
// @Produce json
// @Success 200 {object} types.QueryLiquidityPoolsResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /cosmos/liquidity/v1beta1/pools [get]
func getPools(db *database.Database, s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := s.GetPools()
		if err != nil {
			e := apierrors.New(
				"pools",
				fmt.Sprintf("cannot retrieve pools"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query pools: %w", err),
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, res)
	}
}

// getParams returns the params of liquidity module.
// @Summary Gets params of liquidity module.
// @Tags params
// @ID params
// @Description Gets params of liquidity module.
// @Produce json
// @Success 200 {object} types.QueryParamsResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /cosmos/liquidity/v1beta1/params [get]
func getParams(db *database.Database, s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		res, err := s.GetParams()
		if err != nil {
			e := apierrors.New(
				"params",
				fmt.Sprintf("cannot retrieve params"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve params: %w", err),
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, res)
	}
}

// getSupply returns the total supply.
// @Summary Gets total supply of cosmos-hub
// @Tags supply
// @ID total-supply
// @Description Gets total supply of cosmos hub.
// @Produce json
// @Success 200 {object} types.QueryTotalSupplyResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router / [get]
func getSupply(db *database.Database, s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		res, err := s.GetSupply()
		if err != nil {
			e := apierrors.New(
				"supply",
				fmt.Sprintf("cannot retrieve total supply"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve total supply: %w", err),
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, res)
	}
}

// getNodeInfo returns output of Cosmos's /node_info endpoint.
// @Summary returns output of Cosmos's /node_info endpoint
// @Tags nodeinfo
// @ID node_info
// @Description returns output of Cosmos's /node_info endpoint
// @Produce json
// @Success 200 {object} types.QueryTotalSupplyResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router / [get]
func getNodeInfo(db *database.Database, s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		res, err := s.GetNodeInfo()
		if err != nil {
			e := apierrors.New(
				"node_info",
				fmt.Sprintf("cannot retrieve node_info"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve node_info: %w", err),
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, res)
	}
}
