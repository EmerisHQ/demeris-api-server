package account

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emerishq/emeris-utils/exported/sdktypes"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/emerishq/demeris-api-server/api/apiutils"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/api/router/deps"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

func Register(router *gin.Engine) {
	group := router.Group("/account/:address")
	group.GET("/balance", GetBalancesByAddress)
	group.GET("/stakingbalances", GetDelegationsByAddress)
	group.GET("/unbondingdelegations", GetUnbondingDelegationsByAddress)
	group.GET("/numbers", GetNumbersByAddress)
	group.GET("/tickets", GetUserTickets)
	group.GET("/delegatorrewards/:chain", GetDelegatorRewards)
}

// GetBalancesByAddress returns account of an address.
// @Summary Gets address balance
// @Tags Account
// @ID get-account
// @Description gets address balance
// @Produce json
// @Param address path string true "address to query balance for"
// @Success 200 {object} BalancesResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/balance [get]
func GetBalancesByAddress(c *gin.Context) {
	var res BalancesResponse
	d := deps.GetDeps(c)

	address := c.Param("address")

	balances, err := d.Database.Balances(address)

	if err != nil {
		e := apierrors.New(
			"account",
			fmt.Errorf("cannot retrieve account for address %v", address),
			http.StatusBadRequest,
		).WithLogContext(
			"cannot query database balance for address",
			"address",
			address,
			"error",
			err,
		)
		c.Error(e)
		return
	}

	vd, err := verifiedDenomsMap(d.Database)
	if err != nil {
		e := apierrors.New(
			"account",
			fmt.Errorf("cannot retrieve account for address %v", address),
			http.StatusBadRequest,
		).WithLogContext(
			"cannot query database verified denoms",
			"address",
			address,
			"error",
			err,
		)
		c.Error(e)
		return
	}

	// TODO: get unique chains
	// perhaps we can remove this since there will be another endpoint specifically for fee tokens

	for _, b := range balances {
		res.Balances = append(res.Balances, balanceRespForBalance(
			b,
			vd,
			d.Database.DenomTrace,
		))
	}

	c.JSON(http.StatusOK, res)
}

// What lies ahead is a refactoring operation to ease testing of the algorithm implemented
// to determine whether a given IBC balance is verified or not.
// Since at the time of this commit there isn't a well-formed testing framework for
// api-server, we refactored the algo out, and provided a database querying function type.
// This way we can easily implement table testing for this sensible component, and provide
// fixes to it in a time-sensitive manner.
// This will most probably go away as soon as we have proper testing in place.
type denomTraceFunc func(string, string) (tracelistener.IBCDenomTraceRow, error)

func balanceRespForBalance(rawBalance tracelistener.BalanceRow, vd map[string]bool, dt denomTraceFunc) Balance {
	balance := Balance{
		Address: rawBalance.Address,
		Amount:  rawBalance.Amount,
		OnChain: rawBalance.ChainName,
	}

	verified := vd[rawBalance.Denom]
	baseDenom := rawBalance.Denom

	if rawBalance.Denom[:4] == "ibc/" {
		// is ibc token
		balance.Ibc = IbcInfo{
			Hash: rawBalance.Denom[4:],
		}

		// if err is nil, the ibc denom has a denom trace associated with it
		// so we return it, along with its verified status as well as the complete ibc
		// path

		// otherwise, since we don't touch `verified` and `baseDenom` variables, we stick to the
		// original `ibc/...` denom, which will be unverified by default
		denomTrace, err := dt(rawBalance.ChainName, rawBalance.Denom[4:])
		if err == nil {
			balance.Ibc.Path = denomTrace.Path
			baseDenom = denomTrace.BaseDenom
			verified = vd[denomTrace.BaseDenom]
		}
	}

	balance.Verified = verified
	balance.BaseDenom = baseDenom

	return balance
}

func verifiedDenomsMap(d *database.Database) (map[string]bool, error) {
	chains, err := d.VerifiedDenoms()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]bool)
	for _, cc := range chains {
		for _, vd := range cc {
			ret[vd.Name] = vd.Verified
		}
	}

	return ret, err
}

// GetDelegationsByAddress returns staking account of an address.
// @Summary Gets staking balance
// @Description gets staking balance
// @Tags Account
// @ID get-staking-account
// @Produce json
// @Param address path string true "address to query staking for"
// @Success 200 {object} StakingBalancesResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/stakingbalance [get]
func GetDelegationsByAddress(c *gin.Context) {
	var res StakingBalancesResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	dl, err := d.Database.Delegations(address)

	if err != nil {
		e := apierrors.New(
			"delegations",
			fmt.Errorf("cannot retrieve delegations for address %v", address),
			http.StatusBadRequest,
		).WithLogContext(
			"cannot query database delegations for addresses",
			"address",
			address,
			"error",
			err,
		)
		c.Error(e)

		return
	}

	for _, del := range dl {
		res.StakingBalances = append(res.StakingBalances, StakingBalance{
			ValidatorAddress: del.Validator,
			Amount:           del.Amount,
			ChainName:        del.ChainName,
		})
	}

	c.JSON(http.StatusOK, res)
}

// GetUnbondingDelegationsByAddress returns the unbonding delegations of an address
// @Summary Gets unbonding delegations
// @Description gets unbonding delegations
// @Tags Account
// @ID get-unbonding-delegations-account
// @Produce json
// @Param address path string true "address to query unbonding delegations for"
// @Success 200 {object} UnbondingDelegationsResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/unbondingdelegations [get]
func GetUnbondingDelegationsByAddress(c *gin.Context) {
	var res UnbondingDelegationsResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	unbondings, err := d.Database.UnbondingDelegations(address)

	if err != nil {
		e := apierrors.New(
			"unbonding delegations",
			fmt.Errorf("cannot retrieve unbonding delegations for address %v", address),
			http.StatusBadRequest,
		).WithLogContext(
			"cannot query database unbonding delegations for addresses",
			"address",
			address,
			"error",
			err,
		)
		c.Error(e)

		return
	}

	for _, unbonding := range unbondings {
		res.UnbondingDelegations = append(res.UnbondingDelegations, UnbondingDelegation{
			ValidatorAddress: unbonding.Validator,
			Entries:          unbonding.Entries,
			ChainName:        unbonding.ChainName,
		})
	}

	c.JSON(http.StatusOK, res)
}

// GetDelegatorRewards returns the delegations rewards of an address on a chain
// @Summary Gets delegation rewards
// @Description gets delegation rewards
// @Tags Account
// @ID get-delegation-rewards-account
// @Produce json
// @Param address path string true "address to query delegation rewards for"
// @Param chain path string true "chain to query delegation rewards for"
// @Success 200 {object} DelegatorRewardsResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/delegatorrewards/{chain} [get]
func GetDelegatorRewards(c *gin.Context) {
	var res DelegatorRewardsResponse

	d := deps.GetDeps(c)

	// TODO: add to tracelistener

	address := c.Param("address")
	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		).WithLogContext(
			"cannot retrieve chain",
			"name",
			chainName,
			"error",
			err,
		)
		c.Error(e)

		return
	}

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusInternalServerError,
		).WithLogContext(
			"cannot retrieve chain's sdk-service",
			"name",
			chainName,
			"error",
			err,
		)
		c.Error(e)

		return
	}

	sdkRes, err := client.DelegatorRewards(context.Background(), &sdkutilities.DelegatorRewardsPayload{
		ChainName:    chainName,
		Bech32Prefix: &chain.NodeInfo.Bech32Config.MainPrefix,
		AddresHex:    &address,
	})

	if err != nil {
		e := apierrors.New(
			"chains",
			fmt.Errorf("cannot retrieve delegator rewards from sdk-service"),
			http.StatusInternalServerError,
		).WithLogContext(
			"cannot retrieve delegator rewards from sdk-service",
			"name",
			chainName,
			"error",
			err,
		)
		c.Error(e)

		return
	}

	coinsSlice := func(in []*sdkutilities.Coin) sdktypes.DecCoins {
		ret := sdktypes.DecCoins{}

		for _, c := range in {
			amount, err := sdktypes.NewDecFromStr(c.Amount)
			if err != nil {
				panic(fmt.Errorf("cannot create dec from sdkutilities.Coin amount: %w", err))
			}

			ret = append(ret, sdktypes.DecCoin{
				Denom:  c.Denom,
				Amount: amount,
			})
		}

		return ret
	}

	for _, r := range sdkRes.Rewards {
		res.Rewards = append(res.Rewards, DelegationDelegatorReward{
			ValidatorAddress: r.ValidatorAddress,
			Reward:           coinsSlice(r.Rewards).String(),
		})
	}

	res.Total = coinsSlice(sdkRes.Total).String()

	c.JSON(http.StatusOK, res)
}

// GetNumbersByAddress returns sequence and account number of an address.
// @Summary Gets sequence and account number
// @Description Gets sequence and account number
// @Tags Account
// @ID get-all-numbers-account
// @Produce json
// @Param address path string true "address to query numbers for"
// @Success 200 {object} NumbersResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/numbers [get]
func GetNumbersByAddress(c *gin.Context) {
	var res NumbersResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	dd, err := d.Database.Chains()
	d.Logger.Debugw("chain names", "chain names", dd, "error", err)

	/*
		PSA: do not remove this comment, this is the proper tracelistener-based implementation of this endpoint,
		which will  be used some time in the future as soon as we fix the auth mismatch error.

		dl, err := d.Database.Numbers(address)

		if err != nil {
			e := apierrors.New(
				"numbers",
				fmt.Errorf("cannot retrieve account/sequence numbers for address %v", address),
				http.StatusBadRequest,
			).WithLogContext(
				"cannot query database auth for addresses",
				"address",
				address,
				"error",
				err,
			)

			return
		}*/

	resp, err := fetchNumbers(dd, address)
	if err != nil {
		e := apierrors.New(
			"numbers",
			fmt.Errorf("cannot retrieve account/sequence numbers for address %v", address),
			http.StatusInternalServerError,
		).WithLogContext(
			"cannot query nodes auth for addresses",
			"address",
			address,
			"error",
			err,
		)
		c.Error(e)

		return
	}

	res.Numbers = resp

	c.JSON(http.StatusOK, res)
}

func GetUserTickets(c *gin.Context) {
	d := deps.GetDeps(c)

	address := c.Param("address")

	tickets, err := d.Store.GetUserTickets(address)
	if err != nil {
		e := apierrors.New(
			"tickets",
			fmt.Errorf("cannot retrieve tickets for address %v", address),
			http.StatusBadRequest,
		).WithLogContext(
			"cannot query store for tickets",
			"address",
			address,
			"error",
			err,
		)
		c.Error(e)

		return
	}

	c.JSON(http.StatusOK, UserTicketsResponse{Tickets: tickets})
}

func fetchNumbers(cns []cns.Chain, account string) ([]tracelistener.AuthRow, error) {
	queryGroup, _ := errgroup.WithContext(context.Background())

	results := make([]tracelistener.AuthRow, len(cns))

	for i, chain := range cns {
		iChain := chain
		idx := i
		queryGroup.Go(func() error {
			row, err := apiutils.FetchAccountNumbers(iChain, account)
			if err != nil {
				return fmt.Errorf("unable to get account numbers, %w", err)
			}

			results[idx] = row

			return nil
		})
	}

	if err := queryGroup.Wait(); err != nil {
		return nil, fmt.Errorf("cannot query chains, %w", err)
	}

	for i := 0; i < len(results); i++ {
		if !(results[i].Address == "") {
			continue
		}

		results = append(results[:i], results[i+1:]...)
		i--
	}

	return results, nil
}
