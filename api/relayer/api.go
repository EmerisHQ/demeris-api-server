package relayer

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/apierrors"
	cnsmodels "github.com/emerishq/demeris-backend-models/cns"

	v1 "github.com/allinbits/starport-operator/api/v1"
	"github.com/emerishq/emeris-utils/k8s"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, db *database.Database, i *Informer) {
	rel := router.Group("/relayer")

	rel.GET("/status", getRelayerStatus(i))
	rel.GET("/balance", getRelayerBalance(db, i))
}

// getRelayerStatus returns status of relayer.
// @Summary Gets relayer status
// @Tags Relayer
// @ID relayer-status
// @Description gets relayer status
// @Produce json
// @Success 200 {object} RelayerStatusResponse
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /relayer/status [get]
func getRelayerStatus(ri *Informer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res RelayerStatusResponse

		obj, err := ri.Informer.Lister().Get(k8stypes.NamespacedName{
			Namespace: ri.Namespace,
			Name:      "relayer",
		}.String())

		if err != nil && !errors.Is(err, k8s.ErrNotFound) {
			e := apierrors.New(
				"status",
				fmt.Sprintf("cannot query relayer status"),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot query relayer status: %w", err),
				"obj",
				obj,
			)
			_ = c.Error(e)

			return
		}

		relayer, err := k8s.GetRelayerFromObj(obj)
		if err != nil && !errors.Is(err, k8s.ErrNotFound) {
			e := apierrors.New(
				"status",
				fmt.Sprintf("cannot query relayer status"),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot unstructure relayer status: %w", err),
			)
			_ = c.Error(e)

			return
		}

		res.Running = true

		if errors.Is(err, k8s.ErrNotFound) || relayer.Status.Phase != v1.RelayerPhaseRunning {
			res.Running = false
		}

		c.JSON(http.StatusOK, res)
	}
}

// getRelayerBalance returns the balance of the various relayer accounts.
// @Summary Gets relayer balance for the various relayer accounts
// @Tags Relayer
// @ID relayer-balance
// @Description gets relayer balance for the various relayer accounts
// @Produce json
// @Success 200 {object} RelayerBalances
// @Failure 500,403 {object} apierrors.UserFacingError
// @Router /relayer/balance [get]
func getRelayerBalance(db *database.Database, ri *Informer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var res RelayerBalances

		obj, err := ri.Informer.Lister().Get(k8stypes.NamespacedName{
			Namespace: ri.Namespace,
			Name:      "relayer",
		}.String())

		if err != nil {
			e := apierrors.New(
				"status",
				fmt.Sprintf("cannot query relayer status"),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot query relayer status: %w", err),
				"obj",
				obj,
			)
			_ = c.Error(e)

			return
		}

		relayer, err := k8s.GetRelayerFromObj(obj)
		if err != nil && !errors.Is(err, k8s.ErrNotFound) {
			e := apierrors.New(
				"status",
				fmt.Sprintf("cannot query relayer status"),
				http.StatusInternalServerError,
			).WithLogContext(
				fmt.Errorf("cannot unstructure relayer status: %w", err),
			)
			_ = c.Error(e)
			return
		}

		chains := []string{}
		addresses := []string{}

		for _, cs := range relayer.Status.ChainStatuses {
			chains = append(chains, cs.ID)
			addresses = append(addresses, cs.AccountAddress)
		}

		thresh, err := relayerThresh(ctx, chains, db)
		if err != nil {
			e := apierrors.New(
				"status",
				fmt.Sprintf("cannot retrieve relayer status"),
				http.StatusBadRequest,
			).WithLogContext(
				fmt.Errorf("cannot retrieve relayer status: %w", err),
			)
			_ = c.Error(e)

			return
		}

		for i := 0; i < len(addresses); i++ {
			t, found := thresh[chains[i]]
			if !found {
				continue
			}

			enough, err := enoughBalance(ctx, addresses[i], t, db)
			if err != nil {
				e := apierrors.New(
					"status",
					fmt.Sprintf("cannot retrieve relayer status"),
					http.StatusBadRequest,
				).WithLogContext(
					fmt.Errorf("cannot retrieve relayer status: %w", err),
				)
				_ = c.Error(e)

				return
			}

			res.Balances = append(res.Balances, RelayerBalance{
				Address:       addresses[i],
				ChainName:     chains[i],
				EnoughBalance: enough,
			})

		}

		c.JSON(http.StatusOK, res)
	}
}

func relayerThresh(ctx context.Context, chains []string, db *database.Database) (map[string]cnsmodels.Denom, error) {
	res := map[string]cnsmodels.Denom{}

	for _, cn := range chains {
		chain, err := db.ChainFromChainID(ctx, cn)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // probably the chain isn't enabled yet
			}
			return nil, fmt.Errorf("cannot retrieve chain %s, %w", cn, err)
		}

		res[cn] = chain.RelayerToken()
	}

	return res, nil
}

func enoughBalance(ctx context.Context, address string, denom cnsmodels.Denom, db *database.Database) (bool, error) {
	_, hb, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return false, err
	}

	addrHex := hex.EncodeToString(hb)

	balance, err := db.Balances(ctx, addrHex)
	if err != nil {
		return false, err
	}

	status := false

	for _, bal := range balance {
		if bal.Denom != denom.Name {
			continue
		}

		parsedAmt, err := types.ParseCoinNormalized(bal.Amount)
		if err != nil {
			return false, fmt.Errorf("found relayeramount denom but failed to parse amount, %w", err)
		}

		rThresStr := fmt.Sprintf("%v%s", *denom.MinimumThreshRelayerBalance, parsedAmt.Denom)
		rThresAmt, err := types.ParseCoinNormalized(rThresStr)
		if err != nil {
			return false, fmt.Errorf("cannot ParseCoinNormalized() %s", rThresStr)
		}

		status = parsedAmt.IsGTE(rThresAmt)
		break
	}

	return status, nil
}
