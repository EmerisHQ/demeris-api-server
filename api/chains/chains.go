package chains

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	bech322 "github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	gaia "github.com/cosmos/gaia/v5/app"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/allinbits/demeris-api-server/api/database"
	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/allinbits/demeris-backend-models/cns"
	"github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/cosmos/cosmos-sdk/types/tx"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	mint "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// GetChains returns the list of all the chains supported by demeris.
// @Summary Gets list of supported chains.
// @Tags Chain
// @ID chains
// @Description Gets list of supported chains.
// @Produce json
// @Success 200 {object} chainsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chains [get]
func GetChains(c *gin.Context) {
	var res chainsResponse

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
		res.Chains = append(res.Chains, supportedChain{
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
// @Success 200 {object} chainResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName} [get]
func GetChain(c *gin.Context) {
	var res chainResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
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

	res.Chain = chain

	c.JSON(http.StatusOK, res)
}

// GetChainBech32Config returns bech32 configuration for a chain by specifying its name.
// @Summary Gets chain bech32 configuration by chain name.
// @Tags Chain
// @ID bech32config
// @Description Gets chain bech32 configuration by chain name..
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} bech32ConfigResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/bech32 [get]
func GetChainBech32Config(c *gin.Context) {
	var res bech32ConfigResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
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

	res.Bech32Config = chain.NodeInfo.Bech32Config

	c.JSON(http.StatusOK, res)
}

// GetPrimaryChannelWithCounterparty returns the primary channel of a chain by specifying the counterparty.
// @Summary Gets the channel name that connects two chains.
// @Tags Chain
// @ID counterparty
// @Description Gets the channel name that connects two chains.
// @Param chainName path string true "chain name"
// @Param counterparty path string true "counterparty chain name"
// @Produce json
// @Success 200 {object} primaryChannelResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/primary_channel/{counterparty} [get]
func GetPrimaryChannelWithCounterparty(c *gin.Context) {
	var res primaryChannelResponse

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

	res.Channel = primaryChannel{
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
// @Success 200 {object} primaryChannelsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/primary_channel [get]
func GetPrimaryChannels(c *gin.Context) {
	var res primaryChannelsResponse

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
		res.Channels = append(res.Channels, primaryChannel{
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
// @Success 200 {object} verifiedTraceResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/denom/verify_trace/{hash} [get]
func VerifyTrace(c *gin.Context) {
	var res verifiedTraceResponse

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
		var trace trace

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
			if errors.As(err, &database.ErrNoMatchingChannel{}) {
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
					fmt.Errorf("failed querying for %s", hash),
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

		primaryChannelInfo, err := d.Database.PrimaryChannelCounterparty(chainName, nextChain)

		if err != nil {
			e := deps.NewError(
				"denom/verify-trace",
				fmt.Errorf("failed to get primary channel for %s", hash),
				http.StatusBadRequest,
			)

			d.WriteError(c, e,
				"cannot query primary channel information",
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

		if primaryChannelInfo.ChannelName != channel {

			// save this for the error message when the cause pr is merged
			// e := deps.NewError(
			// 	"denom/verify-trace",
			// 	fmt.Errorf("%s : not primary channel for chain %s- expecting %s got %s", hash, chainName, primaryChannelInfo, channel),
			// 	http.StatusBadRequest,
			// )

			// d.WriteError(c, e,
			// 	"not primary channel",
			// 	"id",
			// 	e.ID,
			// 	"hash",
			// 	hash,
			// 	"channel",
			// 	channel,
			// 	"chain",
			// 	chainName,
			// 	"err",
			// 	err,
			// )

			res.VerifiedTrace.Verified = false

			c.JSON(http.StatusOK, res)
			return
		}
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
// @Success 200 {object} statusResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/status [get]
func GetChainStatus(c *gin.Context) {
	var res statusResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	cbt, err := d.Database.ChainLastBlock(chainName)
	if err != nil {
		e := deps.NewError(
			"status",
			fmt.Errorf("cannot retrieve chain status for %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chain last block time",
			"cannot retrieve chain last block time",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	chain, err := d.Database.Chain(chainName)
	if err != nil {
		e := deps.NewError(
			"status",
			fmt.Errorf("cannot retrieve chain status for %v", chainName),
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

	d.Logger.Debugw("last block time", "chain", chainName, "time", cbt, "threshold_for_chain", chain.ValidBlockThresh.Duration())

	if time.Now().Sub(cbt.BlockTime) > chain.ValidBlockThresh.Duration() {
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
// @Success 200 {object} types.QueryTotalSupplyResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/supply [get]
func GetChainSupply(c *gin.Context) {
	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcPort), grpc.WithInsecure())
	if err != nil {
		e := deps.NewError(
			"supply",
			fmt.Errorf("unable to connect to grpc server for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot connect to grpc",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	bankQuery := bank.NewQueryClient(grpcConn)

	suppRes, err := bankQuery.TotalSupply(context.Background(), &bank.QueryTotalSupplyRequest{})
	if err != nil {
		e := deps.NewError(
			"supply",
			fmt.Errorf("unable to query supply for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"unable to query supply",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, suppRes)
}

// GetChainTx returns the tx info of a given chain.
// @Summary Gets tx info of a given tx.
// @Tags Chain
// @ID tx info
// @Description Gets tx info of a given tx.
// @Param chainName path string true "chain name"
// @Param tx path string true "tx"
// @Produce json
// @Success 200 {object} tx.GetTxResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/txs/{txhash} [get]
func GetChainTx(c *gin.Context) {
	d := deps.GetDeps(c)

	chainName := c.Param("chain")
	txHash := c.Param("tx")

	grpcConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", chainName, grpcPort),
		grpc.WithInsecure(),
	)

	if err != nil {
		e := deps.NewError(
			"tx info",
			fmt.Errorf("unable to connect to grpc server for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot connect to grpc",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	defer grpcConn.Close()

	txClient := tx.NewServiceClient(grpcConn)

	grpcRes, err := txClient.GetTx(context.Background(), &tx.GetTxRequest{Hash: txHash})
	if err != nil {
		e := deps.NewError(
			"tx info",
			fmt.Errorf("unable to fetch tx %s for chain %v", txHash, chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot connect to grpc",
			"id",
			e.ID,
			"name",
			chainName,
			"tx",
			txHash,
			"error",
			err,
		)

		return
	}

	cdc, _ := gaia.MakeCodecs()
	bz, err := cdc.MarshalJSON(grpcRes)

	if err != nil {
		e := deps.NewError(
			"tx info",
			fmt.Errorf("unable to unmarshal tx %s for chain %v", txHash, chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot write tx to response",
			"id",
			e.ID,
			"name",
			chainName,
			"tx",
			txHash,
			"error",
			err,
		)

		return
	}

	c.Data(http.StatusOK, "application/json", bz)
}

// GetNumbersByAddress returns sequence and account number of an address.
// @Summary Gets sequence and account number
// @Description Gets sequence and account number
// @Tags Account
// @ID get-numbers-account
// @Produce json
// @Param address path string true "address to query numbers for"
// @Success 200 {object} numbersResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/numbers/{address} [get]
func GetNumbersByAddress(c *gin.Context) {
	d := deps.GetDeps(c)

	address := c.Param("address")
	chainName := c.Param("chain")

	chainInfo, err := d.Database.Chain(chainName)
	if err != nil {
		e := deps.NewError(
			"numbers",
			fmt.Errorf("cannot retrieve chain data for chain %s", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query chain info for address",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
			"chain",
			chainName,
		)

		return
	}

	resp, err := fetchNumbers(chainInfo, address)
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

	c.JSON(http.StatusOK, numbersResponse{Numbers: resp})
}

func fetchNumbers(chain cns.Chain, account string) (tracelistener.AuthRow, error) {
	accBytes, err := hex.DecodeString(account)
	if err != nil {
		return tracelistener.AuthRow{}, fmt.Errorf("cannot decode hex bytes from account string")
	}

	cdc, _ := simapp.MakeCodecs()

	addr, err := bech322.ConvertAndEncode(chain.NodeInfo.Bech32Config.PrefixAccount, accBytes)
	if err != nil {
		return tracelistener.AuthRow{}, fmt.Errorf("cannot encode bytes to %s acc address, %w", chain.ChainName, err)
	}

	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chain.ChainName, grpcPort), grpc.WithInsecure())
	if err != nil {
		return tracelistener.AuthRow{}, err
	}

	authQuery := types.NewQueryClient(grpcConn)
	resp, err := authQuery.Account(context.Background(), &types.QueryAccountRequest{
		Address: addr,
	})

	if status.Code(err) == codes.NotFound {
		return tracelistener.AuthRow{}, nil
	}

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return tracelistener.AuthRow{}, nil
		}

		return tracelistener.AuthRow{}, fmt.Errorf("cannot query account, %w", err)
	}

	// get a baseAccount
	var accountI types.AccountI
	if err := cdc.UnpackAny(resp.Account, &accountI); err != nil {
		return tracelistener.AuthRow{}, err
	}

	result := tracelistener.AuthRow{
		TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
			ChainName: chain.ChainName,
		},
		Address:        addr,
		SequenceNumber: accountI.GetSequence(),
		AccountNumber:  accountI.GetAccountNumber(),
	}

	return result, nil
}

// GetInflation returns the inflation of a specific chain
// @Summary Gets the inflation of a chain
// @Description Gets inflation
// @Tags Chain
// @ID get-inflation
// @Produce json
// @Success 200 {object} inflationResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/mint/inflation [get]
func GetInflation(c *gin.Context) {

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcPort), grpc.WithInsecure())
	if err != nil {
		e := deps.NewError(
			"mint/inflation",
			fmt.Errorf("unable to connect to grpc server for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot connect to grpc",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	mintQuery := mint.NewQueryClient(grpcConn)

	queryInflationRes, err := mintQuery.Inflation(context.Background(), &mint.QueryInflationRequest{})

	if err != nil {
		e := deps.NewError(
			"mint/inflation",
			fmt.Errorf("unable to query inflation for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"unable to query inflation",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, queryInflationRes)
}

// GetMintParams returns the minting parameters of a specific chain
// @Summary Gets the minting params of a chain
// @Description Gets minting params
// @Tags Chain
// @ID get-mint-params
// @Produce json
// @Success 200 {object} paramsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/mint/params [get]
func GetMintParams(c *gin.Context) {

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcPort), grpc.WithInsecure())
	if err != nil {
		e := deps.NewError(
			"mint/params",
			fmt.Errorf("unable to connect to grpc server for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot connect to grpc",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	mintQuery := mint.NewQueryClient(grpcConn)

	queryParamsRes, err := mintQuery.Params(context.Background(), &mint.QueryParamsRequest{})

	if err != nil {
		e := deps.NewError(
			"mint/params",
			fmt.Errorf("unable to query params for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"unable to query params",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, queryParamsRes)
}

// GetAnnualProvisions returns the annual provisions of a specific chain
// @Summary Gets the annual provisions of a chain
// @Description Gets annual provisions
// @Tags Chain
// @ID get-annual-provisions
// @Produce json
// @Success 200 {object} annualProvisionsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/mint/annual_provisions [get]
func GetAnnualProvisions(c *gin.Context) {

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcPort), grpc.WithInsecure())
	if err != nil {
		e := deps.NewError(
			"mint/annual-provisions",
			fmt.Errorf("unable to connect to grpc server for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot connect to grpc",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	mintQuery := mint.NewQueryClient(grpcConn)

	queryAnnualProvisionsRes, err := mintQuery.AnnualProvisions(context.Background(), &mint.QueryAnnualProvisionsRequest{})

	if err != nil {
		e := deps.NewError(
			"mint/annual-provisions",
			fmt.Errorf("unable to query annual provisions for chain %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"unable to query annual provisions",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, queryAnnualProvisionsRes)
}
