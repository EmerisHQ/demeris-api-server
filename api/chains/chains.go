package chains

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	// needed for swagger gen
	_ "encoding/json"

	"github.com/gin-gonic/gin"

	"github.com/emerishq/demeris-api-server/api/apiutils"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/ginutils"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/demeris-backend-models/cns"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

// GetChains returns the list of all the chains supported by demeris.
// @Summary Gets list of supported chains.
// @Tags Chain
// @ID chains
// @Description Gets list of supported chains.
// @Produce json
// @Success 200 {object} ChainsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chains [get]
func GetChains(c *gin.Context) {
	var res ChainsResponse

	d := deps.GetDeps(c)

	chains, err := d.Database.SimpleChains()

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chains",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for _, cc := range chains {
		res.Chains = append(res.Chains, SupportedChain{
			ChainName:   cc.ChainName,
			DisplayName: cc.DisplayName,
			Logo:        cc.Logo,
		})
	}

	c.JSON(http.StatusOK, res)
}

// GetChain returns chain information by specifying its name.
// @Summary Gets chain by name.
// @Tags Chain
// @ID chain
// @Description Gets chain by name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} ChainResponse
// @Failure 500,400 {object} deps.Error
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
// @Failure 500,403 {object} deps.Error
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
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/primary_channel/{counterparty} [get]
func GetPrimaryChannelWithCounterparty(c *gin.Context) {
	var res PrimaryChannelResponse

	d := deps.GetDeps(c)
	chainName := c.Param("chain")
	counterparty := c.Param("counterparty")
	chain, err := d.Database.PrimaryChannelCounterparty(chainName, counterparty)
	if err != nil {
		e := deps.NewError(
			"primarychannel",
			fmt.Errorf("cannot retrieve primary channel between %v and %v", chainName, counterparty),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"counterparty",
			counterparty,
			"error",
			err,
		)

		return
	}

	res.Channel = PrimaryChannel{
		Counterparty: counterparty,
		ChannelName:  chain.ChannelName,
	}

	c.JSON(http.StatusOK, res)
}

// GetPrimaryChannels returns the primary channels of a chain.
// @Summary Gets the channel mapping of a chain with all the other chains it is connected to.
// @Tags Chain
// @ID channels
// @Description Gets the channel mapping of a chain with all the other chains it is connected to.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} PrimaryChannelsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/primary_channel [get]
func GetPrimaryChannels(c *gin.Context) {
	var res PrimaryChannelsResponse

	d := deps.GetDeps(c)
	chainName := c.Param("chain")
	chain, err := d.Database.PrimaryChannels(chainName)
	if err != nil {
		e := deps.NewError(
			"primarychannel",
			fmt.Errorf("cannot retrieve primary channels for %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

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

// VerifyTrace verifies that a trace hash is valid against a chain name.
// @Summary Verifies that a trace hash is valid against a chain name.
// @Tags Chain
// @ID verifyTrace
// @Description Verifies that a trace hash is valid against a chain name.
// @Param chainName path string true "chain name"
// @Param hash path string true "trace hash"
// @Produce json
// @Success 200 {object} VerifiedTraceResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/denom/verify_trace/{hash} [get]
func VerifyTrace(c *gin.Context) {
	var res VerifiedTraceResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")
	hash := c.Param("hash")

	hash = strings.ToLower(hash)

	res.VerifiedTrace.IbcDenom = fmt.Sprintf("ibc/%s", hash)

	denomTrace, err := d.Database.DenomTrace(chainName, hash)

	if err != nil {

		cause := fmt.Sprintf("token hash %v not found on chain %v", hash, chainName)

		d.LogError(
			cause,
			"hash",
			hash,
			"chainName",
			chainName,
			"error",
			err,
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

		d.LogError(
			"invalid denom",
			"hash",
			hash,
			"path",
			res.VerifiedTrace.Path,
			"err",
			cause,
		)

		res.VerifiedTrace.Verified = false
		res.VerifiedTrace.Cause = cause

		c.JSON(http.StatusOK, res)

		return
	}

	chainIDsMap, err := d.Database.ChainIDs()

	if err != nil {

		err = fmt.Errorf("cannot query list of chain ids, %w", err)

		e := deps.NewError(
			"denom/verify-trace",
			fmt.Errorf("cannot query list of chain ids"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query list of chain ids",
			"id",
			e.ID,
			"hash",
			hash,
			"path",
			res.VerifiedTrace.Path,
			"err",
			err,
		)
		return
	}

	nextChain := chainName
	for _, element := range pathsElements {
		// otherwise, check that it has a transfer prefix
		if !strings.HasPrefix(element, "transfer/") {
			cause := fmt.Sprintf("Unsupported path %s", res.VerifiedTrace.Path)

			d.LogError(
				"invalid denom",
				"hash",
				hash,
				"path",
				res.VerifiedTrace.Path,
				"err",
				cause,
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

			d.LogError(
				"cannot check path element during path resolution",
				"hash",
				hash,
				"path",
				res.VerifiedTrace.Path,
				"err",
				fmt.Errorf("cannot find %s in chainIDs map", nextChain),
			)

			res.VerifiedTrace.Verified = false
			res.VerifiedTrace.Cause = "cannot check path element during path resolution"

			c.JSON(http.StatusOK, res)

			return
		}

		channelInfo, err = d.Database.GetIbcChannelToChain(nextChain, channel, chainID)

		if err != nil {
			if errors.As(err, &database.ErrNoDestChain{}) {
				d.LogError(
					err.Error(),
					"hash",
					hash,
					"path",
					res.VerifiedTrace.Path,
					"chain",
					chainName,
				)

				res.VerifiedTrace.Verified = false
				res.VerifiedTrace.Cause = err.Error()

				c.JSON(http.StatusOK, res)
			} else {
				e1 := deps.NewError(
					"denom/verify-trace",
					fmt.Errorf("failed querying for %s, error: %w", hash, err),
					http.StatusBadRequest,
				)

				d.WriteError(c, e1,
					"invalid number of query responses",
					"id",
					e1.ID,
					"hash",
					hash,
				)
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

	nextChainData, err := d.Database.Chain(nextChain)
	if err != nil {

		d.LogError(
			"cannot query chain",
			"hash",
			hash,
			"path",
			res.VerifiedTrace.Path,
			"nextChain",
			nextChain,
			"err",
			err,
		)

		// we did not find any chain with name nextChain
		if errors.Is(err, sql.ErrNoRows) {
			res.VerifiedTrace.Verified = false
			res.VerifiedTrace.Cause = fmt.Sprintf("no chain with name %s found", nextChain)
			c.JSON(http.StatusOK, res)

			return
		}

		e := deps.NewError(
			"denom/verify-trace",
			fmt.Errorf("database error, %w", err),
			http.StatusInternalServerError,
		)

		d.WriteError(c, e,
			fmt.Sprintf("cannot query chain with name %s", nextChain),
			"id",
			e.ID,
			"hash",
			hash,
			"path",
			res.VerifiedTrace.Path,
			"chain",
			chainName,
			"nextChain",
			nextChain,
			"err",
			err,
		)

		return
	}

	cbt, err := d.Database.ChainLastBlock(nextChain)
	if err != nil {
		e := deps.NewError(
			"denom/verify-trace",
			fmt.Errorf("cannot retrieve chain status for %v", nextChain),
			http.StatusInternalServerError,
		)

		d.WriteError(c, e,
			"cannot retrieve chain last block time",
			"id",
			e.ID,
			"hash",
			hash,
			"path",
			res.VerifiedTrace.Path,
			"chainName",
			chainName,
			"nextChain",
			nextChain,
			"error",
			err,
		)

		return
	}

	d.Logger.Debugw("last block time", "chain", nextChain, "time", cbt, "threshold_for_chain", nextChainData.ValidBlockThresh.Duration())

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
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/status [get]
func GetChainStatus(c *gin.Context) {
	var res StatusResponse
	d := deps.GetDeps(c)

	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

	cbt, err := d.Database.ChainLastBlock(chain.ChainName)
	if err != nil {
		res.Online = false
		c.JSON(http.StatusOK, res)
		return
	}

	d.Logger.Debugw("last block time",
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

// GetChainSupply returns the total supply of a given chain.
// @Summary Gets supply of all denoms of a given chain.
// @Tags Chain
// @ID supply
// @Description Gets supply of a given chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} SupplyResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/supply [get]
func GetChainSupply(c *gin.Context) {
	d := deps.GetDeps(c)

	paginationKey, exists := c.GetQuery("key")
	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	payload := &sdkutilities.SupplyPayload{
		ChainName: chain.ChainName,
	}

	if exists {
		payload.PaginationKey = &paginationKey
	}

	sdkRes, err := client.Supply(context.Background(), payload)
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve supply from sdk-service"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve supply from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

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

// GetDenomSupply returns the total supply of a given denom.
// @Summary Gets supply of a denom of a given chain.
// @Tags Chain
// @ID supply
// @Description Gets supply of a given denom.
// @Param chainName path string true "chain name"
// @Param denom path string true "denom name"
// @Produce json
// @Success 200 {object} SupplyResponse
// @Failure 400 {object} deps.Error
// @Router /chain/{chainName}/supply/:denom [get]
func GetDenomSupply(c *gin.Context) {
	d := deps.GetDeps(c)

	denom := c.Param("denom")
	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id", e.ID,
			"name", chain.ChainName,
			"error", err,
		)

		return
	}

	payload := &sdkutilities.SupplyDenomPayload{
		ChainName: chain.ChainName,
		Denom:     &denom,
	}

	sdkRes, err := client.SupplyDenom(context.Background(), payload)
	if err != nil || len(sdkRes.Coins) != 1 { // Expected exactly one response
		cause := fmt.Errorf("cannot retrieve supply for chain: %s - denom: %s from sdk-service", chain.ChainName, denom)
		if len(sdkRes.Coins) != 1 {
			cause = fmt.Errorf("expected 1 denom for chain: %s - denom: %s, found %v", chain.ChainName, denom, sdkRes.Coins)
		}
		e := deps.NewError(
			"chains",
			cause,
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve denom supply from sdk-service",
			"id", e.ID,
			"chain name", chain.ChainName,
			"denom name", denom,
			"error", err,
		)

		return
	}

	res := SupplyResponse{Supply: []Coin{{Denom: denom, Amount: sdkRes.Coins[0].Amount}}}
	c.JSON(http.StatusOK, res)
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
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/txs/{txhash} [get]
func GetChainTx(c *gin.Context) {
	d := deps.GetDeps(c)

	txHash := c.Param("tx")
	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.QueryTx(context.Background(), &sdkutilities.QueryTxPayload{
		ChainName: chain.ChainName,
		Hash:      txHash,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve tx from sdk-service, %w", err),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve tx from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, sdkRes)
}

// GetNumbersByAddress returns sequence and account number of an address.
// @Summary Gets sequence and account number
// @Description Gets sequence and account number
// @Tags Account
// @ID get-numbers-account
// @Produce json
// @Param address path string true "address to query numbers for"
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/numbers/{address} [get]
func GetNumbersByAddress(c *gin.Context) {
	d := deps.GetDeps(c)

	address := c.Param("address")
	chainInfo := ginutils.GetValue[cns.Chain](c, ChainContextKey)

	resp, err := apiutils.FetchAccountNumbers(chainInfo, address)
	if err != nil {
		e := deps.NewError(
			"numbers",
			fmt.Errorf("cannot retrieve account/sequence numbers for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query nodes auth for address",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
			"chain",
			chainInfo,
		)

		return
	}

	c.JSON(http.StatusOK, NumbersResponse{Numbers: resp})
}

// GetInflation returns the inflation of a specific chain
// @Summary Gets the inflation of a chain
// @Description Gets inflation
// @Tags Chain
// @ID get-inflation
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/mint/inflation [get]
func GetInflation(c *gin.Context) {
	d := deps.GetDeps(c)

	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.MintInflation(context.Background(), &sdkutilities.MintInflationPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve inflation from sdk-service"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve inflation from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintInflation)
}

// GetStakingParams returns the staking parameters of a specific chain
// @Summary Gets the staking parameters of a chain
// @Description Gets staking parameters
// @Tags Chain
// @ID get-staking-params
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 400 {object} deps.Error
// @Router /chain/{chainName}/staking/params [get]
func GetStakingParams(c *gin.Context) {
	d := deps.GetDeps(c)

	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.StakingParams(context.Background(), &sdkutilities.StakingParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve staking params from sdk-service"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve staking params from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.StakingParams)
}

// GetStakingPool returns the staking pool of a specific chain
// @Summary Gets the staking pool of a chain
// @Description Gets staking pool
// @Tags Chain
// @ID get-staking-pool
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 400 {object} deps.Error
// @Router /chain/{chainName}/staking/pool [get]
func GetStakingPool(c *gin.Context) {
	d := deps.GetDeps(c)

	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.StakingPool(context.Background(), &sdkutilities.StakingPoolPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve staking pool from sdk-service"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve staking pool from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.StakingPool)
}

// GetMintParams returns the minting parameters of a specific chain
// @Summary Gets the minting params of a chain
// @Description Gets minting params
// @Tags Chain
// @ID get-mint-params
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/mint/params [get]
func GetMintParams(c *gin.Context) {
	d := deps.GetDeps(c)

	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.MintParams(context.Background(), &sdkutilities.MintParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve mint params from sdk-service"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve mint params from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintParams)
}

// GetAnnualProvisions returns the annual provisions of a specific chain
// @Summary Gets the annual provisions of a chain
// @Description Gets annual provisions
// @Tags Chain
// @ID get-annual-provisions
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/mint/annual_provisions [get]
func GetAnnualProvisions(c *gin.Context) {
	d := deps.GetDeps(c)

	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)
	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.MintAnnualProvision(context.Background(), &sdkutilities.MintAnnualProvisionPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve mint annual provision from sdk-service"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve mint annual provision from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintAnnualProvision)
}

// GetEpochProvisions returns the epoch provisions of a specific chain
// @Summary Gets the epoch provisions of a chain
// @Description Gets epoch provisions
// @Tags Chain
// @ID get-epoch-provisions
// @Produce json
// @Success 200 {object} json.RawMessage
// @Failure 400 {object} deps.Error
// @Router /chain/{chainName}/mint/epoch_provisions [get]
func GetEpochProvisions(c *gin.Context) {
	d := deps.GetDeps(c)

	chain := ginutils.GetValue[cns.Chain](c, ChainContextKey)

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.MintEpochProvisions(context.Background(), &sdkutilities.MintEpochProvisionsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve mint epoch provisions from sdk-service"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve mint epoch provisions from sdk-service",
			"id",
			e.ID,
			"name",
			chain.ChainName,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, gin.MIMEJSON, sdkRes.MintEpochProvisions)
}
