package chains

import (
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/gin-gonic/gin"
)

type ChainAPI struct {
	cacheBackend CacheBackend
	app          App
}

func New(c CacheBackend, app App) *ChainAPI {
	return &ChainAPI{
		cacheBackend: c,
		app:          app,
	}
}

func Register(router *gin.Engine, db *database.Database, cacheBackend CacheBackend, sdkServiceClients sdkservice.SDKServiceClients, app App) {
	chainAPI := New(cacheBackend, app)

	router.Group("/chains").
		GET("", GetChains(db)).
		GET("/status", GetChainsStatuses(db)).
		GET("/fee/addresses", GetFeeAddresses(db))

	chain := router.Group("/chain/:chain")

	chain.GET("/denom/verify_trace/:hash", VerifyTrace(db))

	chain.Group("").
		Use(RequireChainEnabled("chain", db)).
		GET("/primary_channels", GetPrimaryChannels(db)).
		GET("/primary_channel/:counterparty", GetPrimaryChannelWithCounterparty(db)).
		GET("/validators", GetValidators(db, cacheBackend))

	chain.Use(GetChainMiddleware("chain", db)).
		GET("", GetChain).
		GET("/bech32", GetChainBech32Config).
		GET("/status", GetChainStatus(db)).
		GET("/supply", GetChainSupply(sdkServiceClients)).
		GET("/supply/:denom", GetDenomSupply(sdkServiceClients)).
		GET("/txs/:tx", GetChainTx(sdkServiceClients)).
		GET("/numbers/:address", GetNumbersByAddress(sdkServiceClients)).
		GET("/mint/inflation", GetInflation(sdkServiceClients)).
		GET("/mint/params", GetMintParams(sdkServiceClients)).
		GET("/mint/annual_provisions", GetAnnualProvisions(sdkServiceClients)).
		GET("/mint/epoch_provisions", GetEpochProvisions(sdkServiceClients)).
		GET("/staking/params", GetStakingParams(sdkServiceClients)).
		GET("/apr", chainAPI.GetStakingAPR).
		GET("/staking/pool", GetStakingPool(sdkServiceClients)).
		GET("/distribution/params", GetDistributionParams(sdkServiceClients)).
		GET("/budget/params", GetBudgetParams(sdkServiceClients))

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
		ctx := c.Request.Context()
		chainName := c.Param(chainNameParamKey)

		chain, err := db.Chain(ctx, chainName)
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
		ctx := c.Request.Context()
		chainName := c.Param(chainNameParamKey)

		if exists, err := db.ChainExists(ctx, chainName); err != nil || !exists {
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
