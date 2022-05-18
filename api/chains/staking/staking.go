package staking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type Staking struct {
	sdkClient SDKClient
}

func New(sdk SDKClient) *Staking {
	return &Staking{
		sdkClient: sdk,
	}
}

type StakingPoolResponse struct {
	Pool struct {
		NotBondedTokens string `json:"not_bonded_tokens"`
		BondedTokens    string `json:"bonded_tokens"`
	} `json:"pool"`
}

type StakingParamsResponse struct {
	Params struct {
		UnbondingTime     int64  `json:"unbonding_time"`
		MaxValidators     int64  `json:"max_validators"`
		MaxEntries        int64  `json:"max_entries"`
		HistoricalEntries int64  `json:"historical_entries"`
		BondDenom         string `json:"bond_denom"`
	} `json:"params"`
}

func (s *Staking) APR(ctx context.Context, chain cns.Chain) (float64, error) {
	//-----------------------------------------
	// 1- get bonded tokens

	stakingPoolRes, err := s.sdkClient.StakingPool(ctx, &sdkutilities.StakingPoolPayload{
		ChainName: chain.ChainName,
	})
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot retrieve staking pool from sdk-service",
			http.StatusBadRequest,
		)
	}
	var stakingPoolData StakingPoolResponse
	err = json.Unmarshal(stakingPoolRes.StakingPool, &stakingPoolData)
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot unmarshal staking pool",
			http.StatusBadRequest,
		)
	}

	bondedTokens, err := sdktypes.NewDecFromStr(stakingPoolData.Pool.BondedTokens)
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			fmt.Sprintf("cannot convert bonded_tokens to sdktypes.Dec"),
			http.StatusBadRequest,
		)
	}

	//-----------------------------------------
	// 2- get supply

	stakingParamsRes, err := s.sdkClient.StakingParams(ctx, &sdkutilities.StakingParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot retrieve staking params from sdk-service",
			http.StatusBadRequest,
		)
	}

	var stakingParamsData StakingParamsResponse
	err = json.Unmarshal(stakingParamsRes.StakingParams, &stakingParamsData)
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot unmarshal staking params",
			http.StatusBadRequest,
		)
	}

	denomSupplyRes, err := s.sdkClient.SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
		ChainName: chain.ChainName,
		Denom:     &stakingParamsData.Params.BondDenom,
	})
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot retrieve supply denom from sdk-service",
			http.StatusBadRequest,
		)
	}
	if len(denomSupplyRes.Coins) != 1 { // Expected exactly one response
		return 0, apierrors.New("chains",
			fmt.Sprintf("expected 1 denom for chain: %s - denom: %s, found %d",
				chain.ChainName, stakingParamsData.Params.BondDenom, len(denomSupplyRes.Coins)),
			http.StatusBadRequest,
		)
	}

	// denomSupplyRes.Coins[0].Amount is of pattern {amount}{denom} Ex: 438926033423uxyz
	// Hence, converting it to type coin to extract amount
	coin, err := sdktypes.ParseCoinNormalized(denomSupplyRes.Coins[0].Amount)
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot convert amount to coin",
			http.StatusBadRequest,
		)
	}
	supply := coin.Amount.ToDec()

	//-----------------------------------------
	// get inflation

	inflationRes, err := s.sdkClient.MintInflation(ctx, &sdkutilities.MintInflationPayload{
		ChainName: chain.ChainName,
	})
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot retrieve inflation from sdk-service",
			http.StatusBadRequest,
		)
	}

	var inflationData struct {
		Inflation string `json:"inflation"`
	}
	err = json.Unmarshal(inflationRes.MintInflation, &inflationData)
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot unmarshal inflation",
			http.StatusBadRequest,
		)
	}

	inflation, err := sdktypes.NewDecFromStr(inflationData.Inflation)
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot convert inflation to sdktypes.Dec",
			http.StatusBadRequest,
		)
	}
	// only 25% of the newly minted tokens are distributed as staking rewards for osmosis
	if strings.ToLower(chain.ChainName) == osmosisChainName {
		inflation = inflation.QuoInt64(4)
	}

	//-----------------------------------------
	// Compute APR

	apr := inflation.Quo(bondedTokens.Quo(supply)).MulInt64(100)
	f, err := strconv.ParseFloat(apr.String(), 64)
	if err != nil {
		return 0, apierrors.Wrap(err, "chains",
			"cannot convert apr to float",
			http.StatusBadRequest,
		)
	}
	return f, nil
}
