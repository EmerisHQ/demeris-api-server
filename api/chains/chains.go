package chains

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/emerishq/demeris-api-server/api/apiutils"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/lib/ginutils"
	"github.com/emerishq/demeris-api-server/lib/stringcache"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/emerishq/emeris-utils/store"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

const (
	aprCacheDuration  = 24 * time.Hour
	aprCachePrefix    = "api-server/chain-aprs"
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"

	ecosystemIncentiveBudget = "budget-ecosystem-incentive"
	devTeamBudget            = "budget-dev-team"
)

// GetChains returns the list of all the chains supported by demeris.
// @Summary Gets list of supported chains.
// @Tags Chain
// @ID chains
// @Description Gets list of supported chains.
// @Produce json
// @Success 200 {object} ChainsResponse
// @Failure 500,400 {object} apierrors.UserFacingError
// @Router /chains [get]
func GetChains(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res ChainsResponse

		chains, err := db.ChainsWithStatus()

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve chains"),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot retrieve chains: %w", err),
			)
			_ = c.Error(e)

			return
		}

		res.Chains = chains
		c.JSON(http.StatusOK, res)
	}
}

// GetChain returns chain information by specifying its name.
// @Summary Gets chain by name.
// @Tags Chain
// @ID chain
// @Description Gets chain by name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} ChainResponse
// @Failure 500,400 {object} apierrors.UserFacingError
// @Router /chain/{chainName} [get]
func GetChain(c *gin.Context) {
	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
	c.JSON(http.StatusOK, ChainResponse{
		Chain: chain,
	})
}

// GetChainBech32Config returns bech32 configuration for a chain by specifying its name.
// @Summary Gets chain bech32 configuration by chain name.
// @Tags Chain
// @ID bech32config
// @Description Gets chain bech32 configuration by chain name..
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} Bech32ConfigResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/bech32 [get]
func GetChainBech32Config(c *gin.Context) {
	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
	c.JSON(http.StatusOK, Bech32ConfigResponse{
		Bech32Config: chain.NodeInfo.Bech32Config,
	})
}

// GetPrimaryChannelWithCounterparty returns the primary channel of a chain by specifying the counterparty.
// @Summary Gets the channel name that connects two chains.
// @Tags Chain
// @ID counterparty
// @Description Gets the channel name that connects two chains.
// @Param chainName path string true "chain name"
// @Param counterparty path string true "counterparty chain name"
// @Produce json
// @Success 200 {object} PrimaryChannelResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/primary_channel/{counterparty} [get]
func GetPrimaryChannelWithCounterparty(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res PrimaryChannelResponse

		chainName := c.Param("chain")
		counterparty := c.Param("counterparty")
		chain, err := db.PrimaryChannelCounterparty(chainName, counterparty)
		if err != nil {
			e := apierrors.New(
				"primarychannel",
				fmt.Sprintf("cannot retrieve primary channel between %v and %v", chainName, counterparty),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve chain: %w", err),
				"name",
				chainName,
				"counterparty",
				counterparty,
			)
			_ = c.Error(e)

			return
		}

		res.Channel = PrimaryChannel{
			Counterparty: counterparty,
			ChannelName:  chain.ChannelName,
		}

		c.JSON(http.StatusOK, res)
	}
}

// GetPrimaryChannels returns the primary channels of a chain.
// @Summary Gets the channel mapping of a chain with all the other chains it is connected to.
// @Tags Chain
// @ID channels
// @Description Gets the channel mapping of a chain with all the other chains it is connected to.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} PrimaryChannelsResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/primary_channel [get]
func GetPrimaryChannels(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res PrimaryChannelsResponse

		chainName := c.Param("chain")
		chain, err := db.PrimaryChannels(chainName)
		if err != nil {
			e := apierrors.New(
				"primarychannel",
				fmt.Sprintf("cannot retrieve primary channels for %v", chainName),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve chain: %w", err),
				"name",
				chainName,
			)
			_ = c.Error(e)

			return
		}

		for _, cc := range chain {
			res.Channels = append(res.Channels, PrimaryChannel{
				Counterparty: cc.Counterparty,
				ChannelName:  cc.ChannelName,
			})
		}

		c.JSON(http.StatusOK, res)
	}
}

// VerifyTrace verifies that a trace hash is valid against a chain name.
// @Summary Verifies that a trace hash is valid against a chain name.
// @Tags Chain
// @ID verifyTrace
// @Description Verifies that a trace hash is valid against a chain name.
// @Param chainName path string true "chain name"
// @Param hash path string true "trace hash, case insensitive"
// @Produce json
// @Success 200 {object} VerifiedTraceResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/denom/verify_trace/{hash} [get]
func VerifyTrace(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res VerifiedTraceResponse

		logger := ginutils.GetValue[*zap.SugaredLogger](c, logging.LoggerKey)

		chainName := c.Param("chain")
		hash := c.Param("hash")

		res.VerifiedTrace.IbcDenom = IBCDenomHash(hash)

		denomTrace, err := db.DenomTrace(chainName, hash)

		if err != nil {
			cause := fmt.Sprintf("token hash %v not found on chain %v", hash, chainName)

			logger.Errorw(
				cause,
				"hash", hash,
				"chainName", chainName,
			)

			res.VerifiedTrace.Verified = false
			res.VerifiedTrace.Cause = cause

			c.JSON(http.StatusOK, res)
			return
		}

		res.VerifiedTrace.Path = denomTrace.Path
		res.VerifiedTrace.BaseDenom = denomTrace.BaseDenom

		pathsElements, err := paths(res.VerifiedTrace.Path)

		if err != nil {

			cause := fmt.Sprintf("unsupported path %s", res.VerifiedTrace.Path)

			logger.Errorw(
				"invalid denom",
				"hash", hash,
				"path", res.VerifiedTrace.Path,
				"err", cause,
			)

			res.VerifiedTrace.Verified = false
			res.VerifiedTrace.Cause = cause

			c.JSON(http.StatusOK, res)

			return
		}

		chainIDsMap, err := db.ChainIDs()

		if err != nil {

			err = fmt.Errorf("cannot query list of chain ids, %w", err)

			e := apierrors.New(
				"denom/verify-trace",
				fmt.Sprintf("cannot query list of chain ids"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query list of chain ids: %w", err),
				"hash",
				hash,
				"path",
				res.VerifiedTrace.Path,
			)
			_ = c.Error(e)
			return
		}

		nextChain := chainName
		for _, element := range pathsElements {
			// otherwise, check that it has a transfer prefix
			if !strings.HasPrefix(element, "transfer/") {
				cause := fmt.Sprintf("Unsupported path %s", res.VerifiedTrace.Path)

				logger.Errorw(
					"invalid denom",
					"hash", hash,
					"path", res.VerifiedTrace.Path,
					"err", cause,
				)

				res.VerifiedTrace.Verified = false
				res.VerifiedTrace.Cause = cause

				c.JSON(http.StatusOK, res)

				return
			}

			channel := strings.TrimPrefix(element, "transfer/")

			var channelInfo cns.IbcChannelsInfo
			var trace Trace

			chainID, ok := chainIDsMap[nextChain]
			if !ok {
				logger.Errorw(
					"cannot check path element during path resolution",
					"hash", hash,
					"path", res.VerifiedTrace.Path,
					"err", fmt.Errorf("cannot find %s in chainIDs map", nextChain),
				)

				res.VerifiedTrace.Verified = false
				res.VerifiedTrace.Cause = "cannot check path element during path resolution"

				c.JSON(http.StatusOK, res)

				return
			}

			channelInfo, err = db.GetIbcChannelToChain(nextChain, channel, chainID)

			if err != nil {
				if errors.As(err, &database.ErrNoDestChain{}) {
					logger.Errorw(
						err.Error(),
						"hash", hash,
						"path", res.VerifiedTrace.Path,
						"chain", chainName,
					)

					res.VerifiedTrace.Verified = false
					res.VerifiedTrace.Cause = err.Error()

					c.JSON(http.StatusOK, res)
				} else {
					e1 := apierrors.New(
						"denom/verify-trace",
						fmt.Sprintf("failed querying for %s, error: %v", hash, err),
						http.StatusBadRequest,
					).WithLogContext(
						fmt.Errorf("invalid number of query responses: %w", err),
						"hash",
						hash,
					)
					_ = c.Error(e1)
				}

				return
			}

			trace.ChainName = channelInfo[0].ChainAName
			trace.CounterpartyName = channelInfo[0].ChainBName
			trace.Channel = channelInfo[0].ChainAChannelID
			trace.Port = "transfer"

			res.VerifiedTrace.Trace = append(res.VerifiedTrace.Trace, trace)

			nextChain = trace.CounterpartyName
		}

		nextChainData, err := db.Chain(nextChain)
		if err != nil {
			logger.Errorw(
				"cannot query chain",
				"hash", hash,
				"path", res.VerifiedTrace.Path,
				"nextChain", nextChain,
				"err", err,
			)

			// we did not find any chain with name nextChain
			if errors.Is(err, sql.ErrNoRows) {
				res.VerifiedTrace.Verified = false
				res.VerifiedTrace.Cause = fmt.Sprintf("no chain with name %s found", nextChain)
				c.JSON(http.StatusOK, res)

				return
			}

			e := apierrors.New(
				"denom/verify-trace",
				fmt.Sprintf("database error, %v", err),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot query chain with name: %w", err),
				"hash",
				hash,
				"path",
				res.VerifiedTrace.Path,
				"chain",
				chainName,
				"nextChain",
				nextChain,
			)
			_ = c.Error(e)

			return
		}

		cbt, err := db.ChainLastBlock(nextChain)
		if err != nil {
			e := apierrors.New(
				"denom/verify-trace",
				fmt.Sprintf("cannot retrieve chain status for %v", nextChain),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot retrieve chain last block time: %w", err),
				"hash",
				hash,
				"path",
				res.VerifiedTrace.Path,
				"chainName",
				chainName,
				"nextChain",
				nextChain,
			)
			_ = c.Error(e)

			return
		}

		logger.Debugw("last block time", "chain", nextChain, "time", cbt, "threshold_for_chain", nextChainData.ValidBlockThresh.Duration())

		if time.Since(cbt.BlockTime) > nextChainData.ValidBlockThresh.Duration() {
			res.VerifiedTrace.Verified = false
			res.VerifiedTrace.Cause = fmt.Sprintf("chain %s status offline", nextChain)
			c.JSON(http.StatusOK, res)

			return
		}

		res.VerifiedTrace.Verified = false

		// set verifiedStatus for base denom on nextChain
		for _, d := range nextChainData.Denoms {
			if denomTrace.BaseDenom == d.Name {
				res.VerifiedTrace.Verified = d.Verified
				break
			}
		}

		c.JSON(http.StatusOK, res)
	}
}

func paths(path string) ([]string, error) {
	numSlash := strings.Count(path, "/")
	if numSlash == 1 {
		return []string{path}, nil
	}

	if numSlash%2 == 0 {
		return nil, fmt.Errorf("malformed path")
	}

	spl := strings.Split(path, "/")

	var paths []string
	pathBuild := ""

	for i, e := range spl {
		if i%2 != 0 {
			pathBuild = pathBuild + "/" + e
			paths = append(paths, pathBuild)
			pathBuild = ""
		} else {
			pathBuild = e
		}
	}

	return paths, nil
}

// GetChainStatus returns the status of a given chain.
// @Summary Gets status of a given chain.
// @Tags Chain
// @ID status
// @Description Gets status of a given chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/status [get]
func GetChainStatus(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res StatusResponse

		logger := ginutils.GetValue[*zap.SugaredLogger](c, logging.LoggerKey)
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		cbt, err := db.ChainLastBlock(chain.ChainName)
		if err != nil {
			res.Online = false
			c.JSON(http.StatusOK, res)
			return
		}

		logger.Debugw("last block time",
			"chain", chain.ChainName,
			"time", cbt,
			"threshold_for_chain", chain.ValidBlockThresh.Duration(),
		)

		if time.Since(cbt.BlockTime) > chain.ValidBlockThresh.Duration() {
			res.Online = false
			c.JSON(http.StatusOK, res)
			return
		}

		res.Online = true

		c.JSON(http.StatusOK, res)
	}
}

// GetChainSupply returns the total supply of a given chain.
// @Summary Gets supply of all denoms of a given chain.
// @Tags Chain
// @ID supply
// @Description Gets supply of a given chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} SupplyResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/supply [get]
func GetChainSupply(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		paginationKey, exists := c.GetQuery("key")
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		payload := &sdkutilities.SupplyPayload{
			ChainName: chain.ChainName,
		}

		if exists {
			payload.PaginationKey = &paginationKey
		}

		sdkRes, err := client.Supply(c.Request.Context(), payload)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve supply from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve supply from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		sup := make([]Coin, 0)

		res := SupplyResponse{Supply: sup, Pagination: Pagination{}}

		if sdkRes.Pagination.NextKey != nil {
			res.Pagination.NextKey = *sdkRes.Pagination.NextKey
		}

		if sdkRes.Pagination.Total != nil {
			res.Pagination.Total = *sdkRes.Pagination.Total
		}

		for _, s := range sdkRes.Coins {
			res.Supply = append(res.Supply, Coin{
				Denom:  s.Denom,
				Amount: s.Amount,
			})
		}

		c.JSON(http.StatusOK, res)
	}
}

// GetDenomSupply returns the total supply of a given denom.
// @Summary Gets supply of a denom of a given chain.
// @Tags Chain
// @ID denom-supply
// @Description Gets supply of a given denom.
// @Param chainName path string true "chain name"
// @Param denom path string true "denom name"
// @Produce json
// @Success 200 {object} SupplyResponse
// @Failure 400 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/supply/:denom [get]
func GetDenomSupply(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		denom := c.Param("denom")
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		payload := &sdkutilities.SupplyDenomPayload{
			ChainName: chain.ChainName,
			Denom:     &denom,
		}

		sdkRes, err := client.SupplyDenom(c.Request.Context(), payload)
		if err != nil || len(sdkRes.Coins) != 1 { // Expected exactly one response
			cause := fmt.Sprintf("cannot retrieve supply for chain: %s - denom: %s from sdk-service", chain.ChainName, denom)
			if sdkRes != nil && len(sdkRes.Coins) != 1 {
				cause = fmt.Sprintf("expected 1 denom for chain: %s - denom: %s, found %v", chain.ChainName, denom, sdkRes.Coins)
			}
			e := apierrors.New(
				"chains",
				cause,
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve denom supply from sdk-service: %w", err),
				"chain name", chain.ChainName,
				"denom name", denom,
			)
			_ = c.Error(e)

			return
		}

		res := SupplyResponse{Supply: []Coin{{Denom: denom, Amount: sdkRes.Coins[0].Amount}}}
		c.JSON(http.StatusOK, res)
	}
}

// GetChainTx returns the tx info of a given chain.
// @Summary Gets tx info of a given tx.
// @Tags Chain
// @ID tx info
// @Description Gets tx info of a given tx.
// @Param chainName path string true "chain name"
// @Param tx path string true "tx"
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/txs/{txhash} [get]
func GetChainTx(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		txHash := c.Param("tx")
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.QueryTx(c.Request.Context(), &sdkutilities.QueryTxPayload{
			ChainName: chain.ChainName,
			Hash:      txHash,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve tx from sdk-service, %v", err),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve tx from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes)
	}
}

// GetNumbersByAddress returns sequence and account number of an address.
// @Summary Gets sequence and account number
// @Description Gets sequence and account number
// @Tags Account
// @ID get-numbers-account
// @Produce json
// @Param address path string true "address to query numbers for"
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/numbers/{address} [get]
func GetNumbersByAddress(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		chainInfo := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		resp, err := apiutils.FetchAccountNumbers(c.Request.Context(), chainInfo, address, sdkServiceClients)
		if err != nil {
			e := apierrors.New(
				"numbers",
				fmt.Sprintf("cannot retrieve account/sequence numbers for address %v", address),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query nodes auth for address: %w", err),
				"address",
				address,
				"chain",
				chainInfo,
			)
			_ = c.Error(e)

			return
		}

		c.JSON(http.StatusOK, NumbersResponse{Numbers: resp})
	}
}

// GetInflation returns the inflation of a specific chain
// @Summary Gets the inflation of a chain
// @Description Gets inflation
// @Tags Chain
// @ID get-inflation
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/mint/inflation [get]
func GetInflation(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.MintInflation(c.Request.Context(), &sdkutilities.MintInflationPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve inflation from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve inflation from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintInflation)
	}
}

// GetStakingParams returns the staking parameters of a specific chain
// @Summary Gets the staking parameters of a chain
// @Description Gets staking parameters
// @Tags Chain
// @ID get-staking-params
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 400 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/staking/params [get]
func GetStakingParams(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.StakingParams(c.Request.Context(), &sdkutilities.StakingParamsPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve staking params from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve staking params from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.StakingParams)
	}
}

// GetStakingPool returns the staking pool of a specific chain
// @Summary Gets the staking pool of a chain
// @Description Gets staking pool
// @Tags Chain
// @ID get-staking-pool
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 400 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/staking/pool [get]
func GetStakingPool(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.StakingPool(c.Request.Context(), &sdkutilities.StakingPoolPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve staking pool from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve staking pool from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.StakingPool)
	}
}

// GetMintParams returns the minting parameters of a specific chain
// @Summary Gets the minting params of a chain
// @Description Gets minting params
// @Tags Chain
// @ID get-mint-params
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/mint/params [get]
func GetMintParams(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.MintParams(c.Request.Context(), &sdkutilities.MintParamsPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve mint params from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve mint params from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintParams)
	}
}

// GetAnnualProvisions returns the annual provisions of a specific chain
// @Summary Gets the annual provisions of a chain
// @Description Gets annual provisions
// @Tags Chain
// @ID get-annual-provisions
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/mint/annual_provisions [get]
func GetAnnualProvisions(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.MintAnnualProvision(c.Request.Context(), &sdkutilities.MintAnnualProvisionPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve mint annual provision from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve mint annual provision from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintAnnualProvision)
	}
}

// GetEpochProvisions returns the epoch provisions of a specific chain
// @Summary Gets the epoch provisions of a chain
// @Description Gets epoch provisions
// @Tags Chain
// @ID get-epoch-provisions
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 400 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/mint/epoch_provisions [get]
func GetEpochProvisions(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.MintEpochProvisions(c.Request.Context(), &sdkutilities.MintEpochProvisionsPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve mint epoch provisions from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve mint epoch provisions from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintEpochProvisions)
	}
}

// GetStakingAPR returns the staking APR of a specific chain
// @Summary Gets the staking APR of a chain
// @Description Gets APR
// @Tags Chain
// @ID get-staking-apr
// @Produce json
// @Success 200 {object} APRResponse
// @Failure 500,400 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/APR [get]
func GetStakingAPR(db *database.Database, s *store.Store, sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := ginutils.GetValue[*zap.SugaredLogger](c, logging.LoggerKey)

		chainName := c.Param("chain")

		aprCache := stringcache.NewStringCache(
			logger,
			stringcache.NewStoreBackend(s),
			aprCacheDuration,
			aprCachePrefix,
			getAPR(c, sdkServiceClients),
		)
		aprString, err := aprCache.Get(c.Request.Context(), chainName, false)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot get APR"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot get APR: %w", err),
				"name",
				chainName,
			)
			_ = c.Error(e)

			return
		}

		apr, err := strconv.ParseFloat(aprString, 64)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot convert apr to float"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot convert apr to float: %w", err),
				"name",
				chainName,
				"APR",
				apr,
			)
			_ = c.Error(e)

			return
		}
		res := APRResponse{APR: apr}
		c.JSON(http.StatusOK, res)
	}
}

func getAPR(c *gin.Context, sdkServiceClients sdkservice.SDKServiceClients) stringcache.HandlerFunc {
	return func(ctx context.Context, key string) (string, error) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return "", fmt.Errorf(e.Error())
		}

		// get number of bonded tokens from staking/pool data
		stakingPoolRes, err := client.StakingPool(c.Request.Context(), &sdkutilities.StakingPoolPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve staking pool from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve staking pool from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		var stakingPoolData StakingPoolResponse
		err = json.Unmarshal(stakingPoolRes.StakingPool, &stakingPoolData)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot unmarshal staking pool"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot unmarshal staking pool: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		bondedTokens, err := sdktypes.NewDecFromStr(stakingPoolData.Pool.BondedTokens)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot convert bonded_tokens to sdktypes.Dec"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot convert bonded_tokens to sdktypes.Dec: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		// apr for crescent is calculated differently as it follows custom inflation schedules
		// apr=(1-budget rate)*(1-tax)*CurrentInflationAmount/Bonded tokens
		if strings.ToLower(chain.ChainName) == crescentChainName {
			return getCrescentAPR(c, chain, bondedTokens, client)
		}

		// get staking coin denom from staking params
		stakingParamsRes, err := client.StakingParams(c.Request.Context(), &sdkutilities.StakingParamsPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve staking params from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve staking params from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		var stakingParamsData StakingParamsResponse
		err = json.Unmarshal(stakingParamsRes.StakingParams, &stakingParamsData)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot unmarshal staking params"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot unmarshal staking params: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		bond_denom := stakingParamsData.Params.BondDenom

		// get supply of staking denom
		payload := &sdkutilities.SupplyDenomPayload{
			ChainName: chain.ChainName,
			Denom:     &bond_denom,
		}

		denomSupplyRes, err := client.SupplyDenom(c.Request.Context(), payload)
		if err != nil || len(denomSupplyRes.Coins) != 1 { // Expected exactly one response
			cause := fmt.Sprintf("cannot retrieve supply for chain: %s - denom: %s from sdk-service", chain.ChainName, bond_denom)
			if denomSupplyRes != nil && len(denomSupplyRes.Coins) != 1 {
				cause = fmt.Sprintf("expected 1 denom for chain: %s - denom: %s, found %v", chain.ChainName, bond_denom, denomSupplyRes.Coins)
			}
			e := apierrors.New(
				"chains",
				cause,
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve denom supply from sdk-service: %w", err),
				"chain name", chain.ChainName,
				"denom name", bond_denom,
			)
			_ = c.Error(e)

			return "", err
		}

		// denomSupplyRes.Coins[0].Amount is of pattern {amount}{denom} Ex: 438926033423uxyz
		// Hence, converting it to type coin to extract amount
		coin, err := sdktypes.ParseCoinNormalized(denomSupplyRes.Coins[0].Amount)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot convert amount to coin"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot convert amount to coin: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		supply := coin.Amount.ToDec()

		// get inflation
		inflationRes, err := client.MintInflation(c.Request.Context(), &sdkutilities.MintInflationPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve inflation from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve inflation from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		var inflationData InflationResponse
		err = json.Unmarshal(inflationRes.MintInflation, &inflationData)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot unmarshal inflation"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot unmarshal inflation: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		inflation, err := sdktypes.NewDecFromStr(inflationData.Inflation)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot convert inflation to sdktypes.Dec"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot convert inflation to sdktypes.Dec: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return "", err
		}

		// only 25% of the newly minted tokens are distributed as staking rewards for osmosis
		if strings.ToLower(chain.ChainName) == osmosisChainName {
			inflation = inflation.QuoInt64(4)
		}

		// calculate staking APR
		apr := inflation.Quo(bondedTokens.Quo(supply)).MulInt64(100)
		return apr.String(), nil
	}
}

// apr=(1-budget rate)*(1-tax)*CurrentInflationAmount/Bonded tokens
func getCrescentAPR(c *gin.Context, chain cns.Chain, bondedTokens sdktypes.Dec, client sdkutilities.Client) (string, error) {
	budgetRate, err := getBudgetRate(c, chain, client)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot get budget rate"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot get budget rate: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return "", err
	}

	tax, err := getTax(c, chain, client)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot get tax"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot get tax: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return "", err
	}

	currentInflationAmount, err := getCurrentInflationAmount(c, chain, client)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot get current inflation amount"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot get current inflation amount: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return "", err
	}

	OneDec := sdktypes.NewDec(1)
	apr := OneDec.Sub(tax).Mul(OneDec.Sub(budgetRate)).Mul(currentInflationAmount).Quo(bondedTokens)
	return apr.String(), nil
}

func getBudgetRate(c *gin.Context, chain cns.Chain, client sdkutilities.Client) (sdktypes.Dec, error) {
	var budgetRate sdktypes.Dec

	budgetParamsResp, err := client.BudgetParams(c.Request.Context(), &sdkutilities.BudgetParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve budget params from sdk-service"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve budget params from sdk-service: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return budgetRate, err
	}

	var budgetParamsData BudgetParamsResponse
	err = json.Unmarshal(budgetParamsResp.BudgetParams, &budgetParamsData)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot unmarshal budget params"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot unmarshal budget params: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return budgetRate, err
	}

	for _, budget := range budgetParamsData.Params.Budgets {
		if budget.Name == ecosystemIncentiveBudget || budget.Name == devTeamBudget {
			rate, err := sdktypes.NewDecFromStr(budget.Rate)
			if err != nil {
				e := apierrors.New(
					"chains",
					fmt.Sprintf("cannot convert budget rate to Dec"),
					http.StatusBadRequest,
				).WithLogContext(
					fmt.Errorf("cannot convert budget rate to Dec: %w", err),
					"name",
					chain.ChainName,
				)
				_ = c.Error(e)

				return budgetRate, err
			}
			budgetRate.Add(rate)
		}
	}

	return budgetRate, nil
}

func getTax(c *gin.Context, chain cns.Chain, client sdkutilities.Client) (sdktypes.Dec, error) {
	var tax sdktypes.Dec

	distributionParamsResp, err := client.DistributionParams(c.Request.Context(), &sdkutilities.DistributionParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve distribution params from sdk-service"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve distribution params from sdk-service: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return tax, err
	}

	var distributionParamsData DistributionParamsResponse
	err = json.Unmarshal(distributionParamsResp.DistributionParams, &distributionParamsData)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot unmarshal distribution params"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot unmarshal distribution params: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return tax, err
	}

	tax, err = sdktypes.NewDecFromStr(distributionParamsData.Params.CommunityTax)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot convert tax to Dec"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot convert tax to Dec: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return tax, err
	}
	return tax, nil
}

func getCurrentInflationAmount(c *gin.Context, chain cns.Chain, client sdkutilities.Client) (sdktypes.Dec, error) {
	currentInflationAmount := sdktypes.NewDec(0)

	mintParamsResp, err := client.MintParams(c.Request.Context(), &sdkutilities.MintParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot retrieve mint params from sdk-service"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot retrieve mint params from sdk-service: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return currentInflationAmount, err
	}

	var mintParamsData CrecentMintParamsResponse
	err = json.Unmarshal(mintParamsResp.MintParams, &mintParamsData)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Sprintf("cannot unmarshal distribution params"),
			http.StatusBadRequest,
		).WithLogContext(
			fmt.Errorf("cannot unmarshal distribution params: %w", err),
			"name",
			chain.ChainName,
		)
		_ = c.Error(e)

		return currentInflationAmount, err
	}

	now := time.Now()
	for _, schedule := range mintParamsData.Params.InflationSchedules {
		StartTime, err := time.Parse(now.String(), schedule.StartTime)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot convert start time to time"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot convert start time to time: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return currentInflationAmount, err
		}
		EndTime, err := time.Parse(now.String(), schedule.EndTime)
		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot convert end time to time"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot convert end time to time: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return currentInflationAmount, err
		}
		if StartTime.After(now) && EndTime.Before(now) {
			currentInflationAmount, err = sdktypes.NewDecFromStr(schedule.Amount)
			if err != nil {
				e := apierrors.New(
					"chains",
					fmt.Sprintf("cannot convert amount to dec"),
					http.StatusBadRequest,
				).WithLogContext(
					fmt.Errorf("cannot convert amount to dec: %w", err),
					"name",
					chain.ChainName,
				)
				_ = c.Error(e)

				return currentInflationAmount, err
			}
			break
		}
	}
	return currentInflationAmount, nil
}

// GetChainsStatuses returns the status of all the enabled chains.
// @Summary Gets status for all enabled chains.
// @Tags Chain
// @ID statuses
// @Description Gets status for all enabled chains.
// @Produce json
// @Success 200 {object} ChainsStatusesResponse
// @Failure 500 {object} apierrors.UserFacingError
// @Router /chains/status [get]
func GetChainsStatuses(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		statuses, err := db.ChainsOnlineStatuses()
		if err != nil {
			e := apierrors.New(
				"chain",
				"cannot retrieve online status for chains",
				http.StatusInternalServerError,
			).WithLogContext(
				err,
			)

			_ = c.Error(e)
			return
		}

		res := NewChainsStatusesResponse(len(statuses))
		for _, s := range statuses {
			res.Chains[s.ChainName] = ChainStatus{
				Online: s.Online,
			}
		}

		c.JSON(http.StatusOK, res)
	}
}

// GetDistributionParams returns the distribution params of a specific chain
// @Summary Gets the ditribution params of a chain
// @Description Gets distribution params https://docs.cosmos.network/main/modules/distribution/
// @Tags Chain
// @ID get-distribution-params
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/distribution/params [get]
func GetDistributionParams(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.DistributionParams(c.Request.Context(), &sdkutilities.DistributionParamsPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve distribution params from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve distribution params from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.DistributionParams)
	}
}

// GetBudgetParams returns the budget params of a specific chain
// @Summary Gets the budget params of a chain
// @Description Gets budget params https://github.com/tendermint/budget/blob/main/x/budget/spec/01_concepts.md
// @Tags Chain
// @ID get-budget-params
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /chain/{chainName}/budget/params [get]
func GetBudgetParams(sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.BudgetParams(c.Request.Context(), &sdkutilities.BudgetParamsPayload{
			ChainName: chain.ChainName,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve budget params from sdk-service"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve budget params from sdk-service: %w", err),
				"name",
				chain.ChainName,
			)
			_ = c.Error(e)

			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.BudgetParams)
	}
}
