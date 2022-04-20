package chains

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/emeris-utils/store"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, db *database.Database, s *store.Store) {
	router.GET("/chains", GetChains(db))
	router.GET("/chains/fee/addresses", GetFeeAddresses(db))
	router.GET("/chains/primary_channels", EstimatePrimaryChannels(db, s))

	chain := router.Group("/chain/:chain")

	chain.GET("/denom/verify_trace/:hash", VerifyTrace(db))

	chain.Group("").
		Use(RequireChainEnabled("chain", db)).
		GET("/primary_channels", GetPrimaryChannels(db)).
		GET("/primary_channel/:counterparty", GetPrimaryChannelWithCounterparty(db)).
		GET("/validators", GetValidators(db, s))

	chain.Group("").
		Use(GetChainMiddleware("chain", db)).
		GET("", GetChain).
		GET("/bech32", GetChainBech32Config).
		GET("/status", GetChainStatus(db)).
		GET("/supply", GetChainSupply).
		GET("/supply/:denom", GetDenomSupply).
		GET("/txs/:tx", GetChainTx).
		GET("/numbers/:address", GetNumbersByAddress).
		GET("/mint/inflation", GetInflation).
		GET("/mint/params", GetMintParams).
		GET("/mint/annual_provisions", GetAnnualProvisions).
		GET("/mint/epoch_provisions", GetEpochProvisions).
		GET("/staking/params", GetStakingParams).
		GET("/apr", GetStakingAPR(db, s)).
		GET("/staking/pool", GetStakingPool)

	chain.Group("/fee").
		GET("", GetFee(db)).
		GET("/address", GetFeeAddress(db)).
		GET("/token", GetFeeToken(db))
}

const (
	ChainContextKey = "chain"
)

// GetChainMiddleware the chain from the database and sets its cns.Chain
// definition into the context.
func GetChainMiddleware(chainNameParamKey string, db *database.Database) gin.HandlerFunc {
	// TODO: pass deps to GetChainMiddleware instead of taking them from context
	return func(c *gin.Context) {
		chainName := c.Param(chainNameParamKey)

		chain, err := db.Chain(chainName)
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
func RequireChainEnabled(chainNameParamKey string, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chainName := c.Param(chainNameParamKey)

		if exists, err := db.ChainExists(chainName); err != nil || !exists {
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
