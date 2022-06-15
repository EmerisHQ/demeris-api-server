package account

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emerishq/emeris-utils/exported/sdktypes"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/emerishq/emeris-utils/store"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/emerishq/demeris-api-server/api/apiutils"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-api-server/lib/fflag"
	"github.com/emerishq/demeris-api-server/lib/ginutils"
	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

const (
	FixSlashedDelegations = "fixslasheddelegations"
)

type AccountAPI struct {
	app App
	// FIXME remove when #801 is done
	db *database.Database
}

func New(app App) *AccountAPI {
	return &AccountAPI{
		app: app,
	}
}

func Register(router *gin.Engine, db *database.Database, s *store.Store, sdkServiceClients sdkservice.SDKServiceClients, app App) {

	accountAPI := New(app)
	// FIXME remove when #801 is done
	accountAPI.db = db
	router.GET("/accounts/:rawaddress", accountAPI.GetAccounts)

	group := router.Group("/account/:address")
	group.GET("/balance", accountAPI.GetBalancesByAddress)
	group.GET("/stakingbalances", accountAPI.GetDelegationsByAddress)
	group.GET("/unbondingdelegations", GetUnbondingDelegationsByAddress(db))
	group.GET("/numbers", GetNumbersByAddress(db, sdkServiceClients))
	group.GET("/tickets", GetUserTickets(db, s))
	group.GET("/delegatorrewards/:chain", GetDelegatorRewards(db, sdkServiceClients))
}

// GetAccounts returns accounts from a raw address
// @Summary Gets accounts' balance, delegation and rewards
// @Tags Account
// @ID get-account
// @Description gets accounts' balance, delegation and rewards
// @Produce json
// @Param raw address path string true "raw address to query balance for"
// @Success 200 {object} AccountsResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /accounts/{rawaddress} [get]
func (a *AccountAPI) GetAccounts(c *gin.Context) {
	ctx := c.Request.Context()
	rawAddress := c.Param("rawaddress")
	addrs, err := a.app.DeriveRawAddress(ctx, rawAddress)
	if err != nil {
		err := apierrors.Wrap(err, "account",
			fmt.Sprintf("cannot derive addresses from raw address %v", rawAddress),
			http.StatusBadRequest,
		)
		_ = c.Error(err)
		return
	}
	balances, err := a.app.Balances(ctx, addrs)
	if err != nil {
		err := apierrors.Wrap(err, "account",
			fmt.Sprintf("cannot retrieve balances for raw address %v", rawAddress),
			http.StatusBadRequest,
		)
		_ = c.Error(err)
		return
	}
	stakingBalances, err := a.app.StakingBalances(ctx, addrs)
	if err != nil {
		err := apierrors.Wrap(err, "account",
			fmt.Sprintf("cannot retrieve staking balances for raw address %v", rawAddress),
			http.StatusBadRequest,
		)
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, AccountsResponse{
		Balances:        balances,
		StakingBalances: stakingBalances,
	})
}

// GetBalancesByAddress returns account of an address.
// @Summary Gets address balance
// @Tags Account
// @ID get-account
// @Description gets address balance
// @Produce json
// @Param address path string true "address to query balance for"
// @Success 200 {object} BalancesResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /account/{address}/balance [get]
func (a *AccountAPI) GetBalancesByAddress(c *gin.Context) {
	ctx := c.Request.Context()

	address := c.Param("address")

	balances, err := a.app.Balances(ctx, []string{address})
	if err != nil {
		err := apierrors.Wrap(err, "account",
			fmt.Sprintf("cannot retrieve balances for address %v", address),
			http.StatusBadRequest,
		)
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, BalancesResponse{Balances: balances})
}

// GetDelegationsByAddress returns staking account of an address.
// @Summary Gets staking balance
// @Description gets staking balance
// @Tags Account
// @ID get-staking-account
// @Produce json
// @Param address path string true "address to query staking for"
// @Success 200 {object} StakingBalancesResponse
// @Failure 500,400 {object} apierrors.UserFacingError
// @Router /account/{address}/stakingbalances [get]
func (a *AccountAPI) GetDelegationsByAddress(c *gin.Context) {
	ctx := c.Request.Context()
	var res StakingBalancesResponse

	address := c.Param("address")
	if fflag.Enabled(c, FixSlashedDelegations) {
		balances, err := a.app.StakingBalances(ctx, []string{address})
		if err != nil {
			err := apierrors.Wrap(err, "delegations",
				fmt.Sprintf("cannot retrieve delegations for address %v", address),
				http.StatusBadRequest,
			)
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, StakingBalancesResponse{StakingBalances: balances})

	} else {

		dl, err := a.db.DelegationsOldResponse(ctx, address)

		if err != nil {
			e := apierrors.New(
				"delegations",
				fmt.Sprintf("cannot retrieve delegations for address %v", address),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query database delegations for addresses: %w", err),
				"address",
				address,
			)
			_ = c.Error(e)

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
}

// GetUnbondingDelegationsByAddress returns the unbonding delegations of an address
// @Summary Gets unbonding delegations
// @Description gets unbonding delegations
// @Tags Account
// @ID get-unbonding-delegations-account
// @Produce json
// @Param address path string true "address to query unbonding delegations for"
// @Success 200 {object} UnbondingDelegationsResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /account/{address}/unbondingdelegations [get]
func GetUnbondingDelegationsByAddress(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var res UnbondingDelegationsResponse

		address := c.Param("address")

		unbondings, err := db.UnbondingDelegations(ctx, []string{address})

		if err != nil {
			e := apierrors.New(
				"unbonding delegations",
				fmt.Sprintf("cannot retrieve unbonding delegations for address %v", address),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query database unbonding delegations for addresses: %w", err),
				"address",
				address,
			)
			_ = c.Error(e)

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
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /account/{address}/delegatorrewards/{chain} [get]
func GetDelegatorRewards(db *database.Database, sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var res DelegatorRewardsResponse

		// TODO: add to tracelistener

		address := c.Param("address")
		chainName := c.Param("chain")

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

			return
		}

		client, e := sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
		if e != nil {
			_ = c.Error(e)
			return
		}

		sdkRes, err := client.DelegatorRewards(c.Request.Context(), &sdkutilities.DelegatorRewardsPayload{
			ChainName:    chainName,
			Bech32Prefix: &chain.NodeInfo.Bech32Config.MainPrefix,
			AddresHex:    &address,
		})

		if err != nil {
			e := apierrors.New(
				"chains",
				fmt.Sprintf("cannot retrieve delegator rewards from sdk-service"),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot retrieve delegator rewards from sdk-service: %w", err),
				"name",
				chainName,
			)
			_ = c.Error(e)

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
}

// GetNumbersByAddress returns sequence and account number of an address.
// @Summary Gets sequence and account number
// @Description Gets sequence and account number
// @Tags Account
// @ID get-all-numbers-account
// @Produce json
// @Param address path string true "address to query numbers for"
// @Success 200 {object} NumbersResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /account/{address}/numbers [get]
func GetNumbersByAddress(db *database.Database, sdkServiceClients sdkservice.SDKServiceClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var res NumbersResponse

		logger := ginutils.GetValue[*zap.SugaredLogger](c, logging.LoggerKey)

		address := c.Param("address")

		dd, err := db.Chains(ctx)
		logger.Debugw("chain names", "chain names", dd, "error", err)

		/*
			PSA: do not remove this comment, this is the proper tracelistener-based implementation of this endpoint,
			which will  be used some time in the future as soon as we fix the auth mismatch error.

			dl, err := db.Numbers(address)

			if err != nil {
				e := apierrors.New(
					"numbers",
					fmt.Sprintf("cannot retrieve account/sequence numbers for address %v", address),
					http.StatusBadRequest,
				).WithLogContext(
					fmt.Errorf("cannot query database auth for addresses: %w", err),
					"address",
					address,
				)

				return
			}*/

		resp, err := fetchNumbers(c.Request.Context(), dd, address, sdkServiceClients)
		if err != nil {
			e := apierrors.New(
				"numbers",
				fmt.Sprintf("cannot retrieve account/sequence numbers for address %v", address),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot query nodes auth for addresses: %w", err),
				"address",
				address,
			)
			_ = c.Error(e)

			return
		}

		res.Numbers = resp

		c.JSON(http.StatusOK, res)
	}
}

func GetUserTickets(db *database.Database, s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		address := c.Param("address")

		tickets, err := s.GetUserTickets(address)
		if err != nil {
			e := apierrors.New(
				"tickets",
				fmt.Sprintf("cannot retrieve tickets for address %v", address),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot query store for tickets: %w", err),
				"address",
				address,
			)
			_ = c.Error(e)

			return
		}

		c.JSON(http.StatusOK, UserTicketsResponse{Tickets: tickets})
	}
}

func fetchNumbers(ctx context.Context, cns []cns.Chain, account string, sdkServiceClients sdkservice.SDKServiceClients) ([]tracelistener.AuthRow, error) {
	queryGroup, _ := errgroup.WithContext(ctx)

	results := make([]tracelistener.AuthRow, len(cns))

	for i, chain := range cns {
		iChain := chain
		idx := i
		queryGroup.Go(func() error {
			row, err := apiutils.FetchAccountNumbers(ctx, iChain, account, sdkServiceClients)
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
