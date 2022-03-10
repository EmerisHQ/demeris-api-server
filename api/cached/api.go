package cached

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/gin-gonic/gin"
	_ "github.com/gravity-devs/liquidity/x/liquidity/types"
)

func Register(router *gin.Engine) {
	group := router.Group("/cached/cosmos")

	group.GET("/liquidity/v1beta1/pools", getPools)
	group.GET("/liquidity/v1beta1/params", getParams)
	group.GET("/bank/v1beta1/supply", getSupply)
	group.GET("/node_info", getNodeInfo)
}

// getPools returns the of all pools.
// @Summary Gets pools info.
// @Tags pools
// @ID pools
// @Description Gets info of all pools.`10
// @Produce json
// @Success 200 {object} types.QueryLiquidityPoolsResponse
// @Failure 500,403 {object} deps.Error
// @Router /cosmos/liquidity/v1beta1/pools [get]
func getPools(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetPools()
	if err != nil {
		e := deps.NewError(
			"pools",
			fmt.Errorf("cannot retrieve pools"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query pools",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, res)
}

// getParams returns the params of liquidity module.
// @Summary Gets params of liquidity module.
// @Tags params
// @ID params
// @Description Gets params of liquidity module.
// @Produce json
// @Success 200 {object} types.QueryParamsResponse
// @Failure 500,403 {object} deps.Error
// @Router /cosmos/liquidity/v1beta1/params [get]
func getParams(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetParams()
	if err != nil {
		e := deps.NewError(
			"params",
			fmt.Errorf("cannot retrieve params"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve params",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, res)
}

// getSupply returns the total supply.
// @Summary Gets total supply of cosmos-hub
// @Tags supply
// @ID total-supply
// @Description Gets total supply of cosmos hub.
// @Produce json
// @Success 200 {object} types.QueryTotalSupplyResponse
// @Failure 500,403 {object} deps.Error
// @Router / [get]
func getSupply(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetSupply()
	if err != nil {
		e := deps.NewError(
			"supply",
			fmt.Errorf("cannot retrieve total supply"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve total supply",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, res)
}

// getNodeInfo returns output of Cosmos's /node_info endpoint.
// @Summary returns output of Cosmos's /node_info endpoint
// @Tags nodeinfo
// @ID node_info
// @Description returns output of Cosmos's /node_info endpoint
// @Produce json
// @Success 200 {object} types.QueryTotalSupplyResponse
// @Failure 500,403 {object} deps.Error
// @Router / [get]
func getNodeInfo(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetNodeInfo()
	if err != nil {
		e := deps.NewError(
			"node_info",
			fmt.Errorf("cannot retrieve node_info"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve node_info",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, res)
}
