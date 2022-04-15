package chains

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, d *deps.Deps) {
	router.GET("/chains", GetChains(d))
	router.GET("/chains/fee/addresses", GetFeeAddresses(d))

	chain := router.Group("/chain/:chain")

	chain.GET("/denom/verify_trace/:hash", VerifyTrace(d))

	chain.Group("").
		Use(RequireChainEnabled("chain", d)).
		GET("/primary_channels", GetPrimaryChannels(d)).
		GET("/primary_channel/:counterparty", GetPrimaryChannelWithCounterparty(d)).
		GET("/validators", GetValidators(d))

	chain.Group("").
		Use(GetChainMiddleware("chain", d)).
		GET("", GetChain).
		GET("/bech32", GetChainBech32Config).
		GET("/status", GetChainStatus(d)).
		GET("/supply", GetChainSupply).
		GET("/supply/:denom", GetDenomSupply).
		GET("/txs/:tx", GetChainTx).
		GET("/numbers/:address", GetNumbersByAddress).
		GET("/mint/inflation", GetInflation).
		GET("/mint/params", GetMintParams).
		GET("/mint/annual_provisions", GetAnnualProvisions).
		GET("/mint/epoch_provisions", GetEpochProvisions).
		GET("/staking/params", GetStakingParams).
		GET("/apr", GetStakingAPR(d)).
		GET("/staking/pool", GetStakingPool)

	chain.Group("/fee").
		GET("", GetFee(d)).
		GET("/address", GetFeeAddress(d)).
		GET("/token", GetFeeToken(d))
}

const (
	ChainContextKey = "chain"
)

// GetChainMiddleware the chain from the database and sets its cns.Chain
// definition into the context.
func GetChainMiddleware(chainNameParamKey string, d *deps.Deps) gin.HandlerFunc {
	// TODO: pass deps to GetChainMiddleware instead of taking them from context
	return func(c *gin.Context) {
		chainName := c.Param(chainNameParamKey)

		chain, err := d.Database.Chain(chainName)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve chain with name %v", chainName),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve chain: %w", err),
				"name",
				chainName,
			)
			_ = c.Error(e)
			c.Abort()
			return
		}

		c.Set(ChainContextKey, chain)
		c.Next()
	}
}

// RequireChainEnabled checks if the chain exists and it's enabled in the database,
// if it's not it returns an error to the user.
func RequireChainEnabled(chainNameParamKey string, d *deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		chainName := c.Param(chainNameParamKey)

		if exists, err := d.Database.ChainExists(chainName); err != nil || !exists {
			if err == nil {
				err = fmt.Errorf("%s chain doesnt exists", chainName)
			}

			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve chain with name %v", chainName),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve chain: %w", err),
				"name", chainName,
			)

			_ = c.Error(e)
			c.Abort()
			return
		}

		c.Next()
	}
}
