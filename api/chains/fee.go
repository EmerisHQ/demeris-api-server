package chains

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/gin-gonic/gin"
)

// GetFee returns the fee average in dollar for the specified chain.
// @Summary Gets average fee in dollar by chain name.
// @Tags Chain
// @ID fee
// @Description Gets average fee in dollar by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} FeeResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/fee [get]
func GetFee(c *gin.Context) {
	var res FeeResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := apierrors.New(
			"fee",
			fmt.Sprintf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chain: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	res = FeeResponse{
		Denoms: chain.FeeTokens(),
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddress returns the fee address for a given chain, looked up by the chain name attribute.
// @Summary Gets address to pay fee for by chain name.
// @Tags Chain
// @ID feeaddress
// @Description Gets address to pay fee for by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} FeeAddressResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/address [get]
func GetFeeAddress(c *gin.Context) {
	var res FeeAddressResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := apierrors.New(
			"feeaddress",
			fmt.Sprintf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chain: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	res = FeeAddressResponse{
		FeeAddress: chain.DemerisAddresses,
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddresses returns the fee address for all chains.
// @Summary Gets all addresses to pay fee for.
// @Tags Chain
// @ID feeaddresses
// @Description Gets all addresses to pay fee for.
// @Produce json
// @Success 200 {object} FeeAddressesResponse
// @Failure 500,403 {object} deps.Error
// @Router /chains/fee/addresses [get]
func GetFeeAddresses(c *gin.Context) {
	var res FeeAddressesResponse

	d := deps.GetDeps(c)

	chains, err := d.Database.Chains()

	if err != nil {
		e := apierrors.New(
			"feeaddress",
			fmt.Sprintf("cannot retrieve chains"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chains: %w", err),
		)
		_ = c.Error(e)

		return
	}

	for _, c := range chains {
		res.FeeAddresses = append(
			res.FeeAddresses,
			FeeAddress{
				ChainName:  c.ChainName,
				FeeAddress: c.DemerisAddresses,
			},
		)
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeToken returns the fee token for a given chain, looked up by the chain name attribute.
// @Summary Gets token used to pay fees by chain name.
// @Tags Chain
// @ID feetoken
// @Description Gets token used to pay fees by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} FeeTokenResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/token [get]
func GetFeeToken(c *gin.Context) {
	var res FeeTokenResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := apierrors.New(
			"feetoken",
			fmt.Sprintf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve chain: %w", err),
			"name",
			chainName,
		)
		_ = c.Error(e)

		return
	}

	for _, cc := range chain.FeeTokens() {
		res.FeeTokens = append(res.FeeTokens, cc)
	}

	c.JSON(http.StatusOK, res)
}
