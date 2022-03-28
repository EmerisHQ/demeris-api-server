package chains

import "github.com/gin-gonic/gin"

func Register(router *gin.Engine) {
	router.GET("/chains", GetChains)
	router.GET("/chains/fee/addresses", GetFeeAddresses)

	chain := router.Group("/chain/:chain")

	chain.GET("", GetChain)
	chain.GET("/denom/verify_trace/:hash", VerifyTrace)
	chain.GET("/bech32", GetChainBech32Config)
	chain.GET("/primary_channels", GetPrimaryChannels)
	chain.GET("/primary_channel/:counterparty", GetPrimaryChannelWithCounterparty)
	chain.GET("/status", GetChainStatus)
	chain.GET("/supply", GetChainSupply)
	chain.GET("/supply/:denom", GetDenomSupply)
	chain.GET("/txs/:tx", GetChainTx)
	chain.GET("/numbers/:address", GetNumbersByAddress)
	chain.GET("/validators", GetValidators)
	chain.GET("/mint/inflation", GetInflation)
	chain.GET("/mint/params", GetMintParams)
	chain.GET("/mint/annual_provisions", GetAnnualProvisions)
	chain.GET("/mint/epoch_provisions", GetEpochProvisions)
	chain.GET("/staking/params", GetStakingParams)
	chain.GET("/staking/pool", GetStakingPool)

	fee := chain.Group("/fee")

	fee.GET("", GetFee)
	fee.GET("/address", GetFeeAddress)
	fee.GET("/token", GetFeeToken)
}
