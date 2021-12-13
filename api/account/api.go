package account

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/allinbits/demeris-api-server/api/apiutils"
	"github.com/allinbits/demeris-api-server/api/database"
	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/allinbits/demeris-api-server/sdkservice"
	"github.com/allinbits/demeris-backend-models/cns"
	"github.com/allinbits/demeris-backend-models/tracelistener"
	sdkutilities "github.com/allinbits/sdk-service-meta/gen/sdk_utilities"
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
// @Success 200 {object} balancesResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/balance [get]
func GetBalancesByAddress(c *gin.Context) {
	var res balancesResponse
	d := deps.GetDeps(c)

	address := c.Param("address")

	balances, err := d.Database.Balances(address)

	if err != nil {
		e := deps.NewError(
			"account",
			fmt.Errorf("cannot retrieve account for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query database balance for address",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)
		return
	}

	vd, err := verifiedDenomsMap(d.Database)
	if err != nil {
		e := deps.NewError(
			"account",
			fmt.Errorf("cannot retrieve account for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query database verified denoms",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)
		return
	}

	// TODO: get unique chains
	// perhaps we can remove this since there will be another endpoint specifically for fee tokens

	for _, b := range balances {
		balance := balance{
			Address: b.Address,
			Amount:  b.Amount,
			OnChain: b.ChainName,
		}

		if b.Denom[:4] == "ibc/" {
			// is ibc token
			balance.Ibc = ibcInfo{
				Hash: b.Denom[4:],
			}

			denomTrace, err := d.Database.DenomTrace(b.ChainName, b.Denom[4:])

			if err != nil {
				e := deps.NewError(
					"account",
					fmt.Errorf("cannot query denom trace for token %v on chain %v", b.Denom, b.ChainName),
					http.StatusBadRequest,
				)

				d.WriteError(c, e,
					"cannot query database balance for address",
					"id",
					e.ID,
					"token",
					b.Denom,
					"chain",
					b.ChainName,
					"error",
					err,
				)

				return
			}
			balance.BaseDenom = denomTrace.BaseDenom
			balance.Ibc.Path = denomTrace.Path
			balance.Verified = vd[denomTrace.BaseDenom]
		} else {
			balance.Verified = vd[b.Denom]
			balance.BaseDenom = b.Denom
		}

		res.Balances = append(res.Balances, balance)
	}

	c.JSON(http.StatusOK, res)
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
// @Success 200 {object} stakingBalancesResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/stakingbalance [get]
func GetDelegationsByAddress(c *gin.Context) {
	var res stakingBalancesResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	dl, err := d.Database.Delegations(address)

	if err != nil {
		e := deps.NewError(
			"delegations",
			fmt.Errorf("cannot retrieve delegations for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query database delegations for addresses",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)

		return
	}

	for _, del := range dl {
		res.StakingBalances = append(res.StakingBalances, stakingBalance{
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
// @Success 200 {object} unbondingDelegationsResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/unbondingdelegations [get]
func GetUnbondingDelegationsByAddress(c *gin.Context) {
	var res unbondingDelegationsResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	unbondings, err := d.Database.UnbondingDelegations(address)

	if err != nil {
		e := deps.NewError(
			"unbonding delegations",
			fmt.Errorf("cannot retrieve unbonding delegations for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query database unbonding delegations for addresses",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)

		return
	}

	for _, unbonding := range unbondings {
		res.UnbondingDelegations = append(res.UnbondingDelegations, unbondingDelegation{
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
// @Success 200 {object} delegatorRewardsResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/delegatorrewards/{chain} [get]
func GetDelegatorRewards(c *gin.Context) {
	var res delegatorRewardsResponse

	d := deps.GetDeps(c)

	// TODO: add to tracelistener

	address := c.Param("address")
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

	client, err := sdkservice.Client(chain.MajorSDKVersion())
	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve sdk-service for version %s with chain name %v", chain.CosmosSDKVersion, chain.ChainName),
			http.StatusInternalServerError,
		)

		d.WriteError(c, e,
			"cannot retrieve chain's sdk-service",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	sdkRes, err := client.DelegatorRewards(context.Background(), &sdkutilities.DelegatorRewardsPayload{
		ChainName:    chainName,
		Bech32Prefix: &chain.NodeInfo.Bech32Config.MainPrefix,
		AddresHex:    &address,
	})

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve delegator rewards from sdk-service"),
			http.StatusInternalServerError,
		)

		d.WriteError(c, e,
			"cannot retrieve delegator rewards from sdk-service",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	coinsSlice := func(in []*sdkutilities.Coin) sdkutilities.Coins {
		ret := sdkutilities.Coins{}

		for _, c := range in {
			ret = append(ret, *c)
		}

		return ret
	}

	for _, r := range sdkRes.Rewards {
		res.Rewards = append(res.Rewards, delegationDelegatorReward{
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
// @Success 200 {object} numbersResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/numbers [get]
func GetNumbersByAddress(c *gin.Context) {
	var res numbersResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	dd, err := d.Database.Chains()
	d.Logger.Debugw("chain names", "chain names", dd, "error", err)

	/*
		PSA: do not remove this comment, this is the proper tracelistener-based implementation of this endpoint,
		which will  be used some time in the future as soon as we fix the auth mismatch error.

		dl, err := d.Database.Numbers(address)

		if err != nil {
			e := deps.NewError(
				"numbers",
				fmt.Errorf("cannot retrieve account/sequence numbers for address %v", address),
				http.StatusBadRequest,
			)

			d.WriteError(c, e,
				"cannot query database auth for addresses",
				"id",
				e.ID,
				"address",
				address,
				"error",
				err,
			)

			return
		}*/

	resp, err := fetchNumbers(dd, address)
	if err != nil {
		e := deps.NewError(
			"numbers",
			fmt.Errorf("cannot retrieve account/sequence numbers for address %v", address),
			http.StatusInternalServerError,
		)

		d.WriteError(c, e,
			"cannot query nodes auth for addresses",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)

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
		e := deps.NewError(
			"tickets",
			fmt.Errorf("cannot retrieve tickets for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query store for tickets",
			"address",
			address,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, userTicketsResponse{Tickets: tickets})
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
