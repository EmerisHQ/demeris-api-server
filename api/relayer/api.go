package relayer

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/allinbits/demeris-api-server/api/database"
	cnsmodels "github.com/allinbits/demeris-backend-models/cns"

	"github.com/allinbits/demeris-api-server/api/apierror"
	"github.com/allinbits/demeris-api-server/api/router/deps"
	"github.com/allinbits/emeris-utils/k8s"
	v1 "github.com/allinbits/starport-operator/api/v1"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	rel := router.Group("/relayer")

	rel.GET("/status", getRelayerStatus)
	rel.GET("/balance", getRelayerBalance)
}

// getRelayerStatus returns status of relayer.
// @Summary Gets relayer status
// @Tags Relayer
// @ID relayer-status
// @Description gets relayer status
// @Produce json
// @Success 200 {object} relayerStatusResponse
// @Failure 500,403 {object} deps.Error
// @Router /relayer/status [get]
func getRelayerStatus(c *gin.Context) {
	var res relayerStatusResponse

	d := deps.GetDeps(c)

	obj, err := d.RelayersInformer.Lister().Get(k8stypes.NamespacedName{
		Namespace: d.KubeNamespace,
		Name:      "relayer",
	}.String())

	if err != nil && !errors.Is(err, k8s.ErrNotFound) {
		e := apierror.New(
			"status",
			fmt.Errorf("cannot query relayer status"),
			http.StatusInternalServerError,
		)

		apierror.WriteError(d.Logger, c, e,
			"cannot query relayer status",
			"id",
			e.ID,
			"error",
			err,
			"obj",
			obj,
		)

		return
	}

	relayer, err := k8s.GetRelayerFromObj(obj)
	if err != nil && !errors.Is(err, k8s.ErrNotFound) {
		e := apierror.New(
			"status",
			fmt.Errorf("cannot query relayer status"),
			http.StatusInternalServerError,
		)

		apierror.WriteError(d.Logger, c, e,
			"cannot unstructure relayer status",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	res.Running = true

	if errors.Is(err, k8s.ErrNotFound) || relayer.Status.Phase != v1.RelayerPhaseRunning {
		res.Running = false
	}

	c.JSON(http.StatusOK, res)
}

// getRelayerBalance returns the balance of the various relayer accounts.
// @Summary Gets relayer balance for the various relayer accounts
// @Tags Relayer
// @ID relayer-balance
// @Description gets relayer balance for the various relayer accounts
// @Produce json
// @Success 200 {object} relayerBalances
// @Failure 500,403 {object} deps.Error
// @Router /relayer/balance [get]
func getRelayerBalance(c *gin.Context) {
	var res relayerBalances

	d := deps.GetDeps(c)

	obj, err := d.RelayersInformer.Lister().Get(k8stypes.NamespacedName{
		Namespace: d.KubeNamespace,
		Name:      "relayer",
	}.String())

	if err != nil {
		e := apierror.New(
			"status",
			fmt.Errorf("cannot query relayer status"),
			http.StatusInternalServerError,
		)

		apierror.WriteError(d.Logger, c, e,
			"cannot query relayer status",
			"id",
			e.ID,
			"error",
			err,
			"obj",
			obj,
		)

		return
	}

	relayer, err := k8s.GetRelayerFromObj(obj)
	if err != nil && !errors.Is(err, k8s.ErrNotFound) {
		e := apierror.New(
			"status",
			fmt.Errorf("cannot query relayer status"),
			http.StatusInternalServerError,
		)

		apierror.WriteError(d.Logger, c, e,
			"cannot unstructure relayer status",
			"id",
			e.ID,
			"error",
			err,
		)

	}

	chains := []string{}
	addresses := []string{}

	for _, cs := range relayer.Status.ChainStatuses {
		chains = append(chains, cs.ID)
		addresses = append(addresses, cs.AccountAddress)
	}

	thresh, err := relayerThresh(chains, d.Database)
	if err != nil {
		e := apierror.New(
			"status",
			fmt.Errorf("cannot retrieve relayer status"),
			http.StatusBadRequest,
		)

		apierror.WriteError(d.Logger, c, e,
			"cannot retrieve relayer status",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for i := 0; i < len(addresses); i++ {
		t, found := thresh[chains[i]]
		if !found {
			continue
		}

		enough, err := enoughBalance(addresses[i], t, d.Database)
		if err != nil {
			e := apierror.New(
				"status",
				fmt.Errorf("cannot retrieve relayer status"),
				http.StatusBadRequest,
			)

			apierror.WriteError(d.Logger, c, e,
				"cannot retrieve relayer status",
				"id",
				e.ID,
				"error",
				err,
			)

			return
		}

		res.Balances = append(res.Balances, relayerBalance{
			Address:       addresses[i],
			ChainName:     chains[i],
			EnoughBalance: enough,
		})

	}

	c.JSON(http.StatusOK, res)
}

func relayerThresh(chains []string, db *database.Database) (map[string]cnsmodels.Denom, error) {
	res := map[string]cnsmodels.Denom{}

	for _, cn := range chains {
		chain, err := db.ChainFromChainID(cn)
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

func enoughBalance(address string, denom cnsmodels.Denom, db *database.Database) (bool, error) {
	_, hb, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return false, err
	}

	addrHex := hex.EncodeToString(hb)

	balance, err := db.Balances(addrHex)
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
