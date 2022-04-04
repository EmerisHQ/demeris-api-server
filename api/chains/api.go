package chains

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/chains", GetChains)
	router.GET("/chains/fee/addresses", GetFeeAddresses)

	chain := router.Group("/chain/:chain")

	chain.GET("/denom/verify_trace/:hash", VerifyTrace)

	chain.Group("").
		Use(RequireChainEnabled("chain")).
		GET("/primary_channels", GetPrimaryChannels).
		GET("/primary_channel/:counterparty", GetPrimaryChannelWithCounterparty).
		GET("/validators", GetValidators)

	chain.Group("").
		Use(GetChainMiddleware("chain")).
		GET("", GetChain).
		GET("/bech32", GetChainBech32Config).
		GET("/status", GetChainStatus).
		GET("/supply", GetChainSupply).
		GET("/supply/:denom", GetDenomSupply).
		GET("/txs/:tx", GetChainTx).
		GET("/numbers/:address", GetNumbersByAddress).
		GET("/mint/inflation", GetInflation).
		GET("/mint/params", GetMintParams).
		GET("/mint/annual_provisions", GetAnnualProvisions).
		GET("/mint/epoch_provisions", GetEpochProvisions).
		GET("/staking/params", GetStakingParams).
		GET("/apr", GetStakingAPR).
		GET("/staking/pool", GetStakingPool)

	chain.Group("/fee").
		GET("", GetFee).
		GET("/address", GetFeeAddress).
		GET("/token", GetFeeToken)
}

const (
	ChainContextKey = "chain"
)

// GetChainMiddleware the chain from the database and sets its cns.Chain
// definition into the context.
func GetChainMiddleware(chainNameParamKey string) gin.HandlerFunc {
	// TODO: pass deps to GetChainMiddleware instead of taking them from context
	return func(c *gin.Context) {
		chainName := c.Param(chainNameParamKey)
		d := deps.GetDeps(c)

		chain, err := d.Database.Chain(chainName)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Errorf("cannot retrieve chain with name %v", chainName),
				http.StatusBadRequest,
			)

			d.WriteError(c, e,
				"cannot retrieve chain",
				"name",
				chainName,
				"error",
				err,
			)

			c.Abort()
			return
		}

		c.Set(ChainContextKey, chain)
		c.Next()
	}
}

// RequireChainEnabled checks if the chain exists and it's enabled in the database,
// if it's not it returns an error to the user.
func RequireChainEnabled(chainNameParamKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		d := deps.GetDeps(c)

		chainName := c.Param(chainNameParamKey)

		if exists, err := d.Database.ChainExists(chainName); err != nil || !exists {
			e := apierrors.New(
				"chains",
				fmt.Errorf("cannot retrieve chain with name %v", chainName),
				http.StatusBadRequest,
			)

			if err == nil {
				err = fmt.Errorf("%s chain doesnt exists", chainName)
			}

			d.WriteError(c, e,
				"cannot retrieve chain",
				"name",
				chainName,
				"error",
				err,
			)

			c.Abort()
			return
		}

		c.Next()
	}
}
