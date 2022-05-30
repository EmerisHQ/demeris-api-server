package usecase

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

func (app *App) StakingAPR(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error) {

	sdkClient, err := app.sdkServiceClients.GetSDKServiceClient(chain.MajorSDKVersion())
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			fmt.Sprintf("cannot retrieve sdk service for version %s", chain.MajorSDKVersion()),
			http.StatusBadRequest,
		)
	}

	//-----------------------------------------
	// 1- get bonded tokens

	stakingPool, err := getStakingPool(ctx, sdkClient, chain.ChainName)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve staking pool from sdk-service",
			http.StatusBadRequest,
		)
	}
	bondedTokens, err := sdktypes.NewDecFromStr(stakingPool.Pool.BondedTokens)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot convert bonded_tokens to sdktypes.Dec",
			http.StatusBadRequest,
		)
	}

	// apr for crescent is calculated differently as it follows custom inflation schedules
	// apr=(1-budget rate)*(1-tax)*CurrentInflationAmount/Bonded tokens
	if strings.ToLower(chain.ChainName) == crescentChainName {
		return app.getCrescentAPR(ctx, sdkClient, chain, bondedTokens)
	}

	//-----------------------------------------
	// 2- get supply

	stakingParams, err := getStakingParams(ctx, sdkClient, chain.ChainName)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve staking params from sdk-service",
			http.StatusBadRequest,
		)
	}

	denomSupplyRes, err := sdkClient.SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
		ChainName: chain.ChainName,
		Denom:     &stakingParams.Params.BondDenom,
	})
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve supply denom from sdk-service",
			http.StatusBadRequest,
		)
	}
	if len(denomSupplyRes.Coins) != 1 { // Expected exactly one response
		return sdktypes.Dec{}, apierrors.New("chains",
			fmt.Sprintf("expected 1 denom for chain: %s - denom: %s, found %d",
				chain.ChainName, stakingParams.Params.BondDenom, len(denomSupplyRes.Coins)),
			http.StatusBadRequest,
		)
	}

	// denomSupplyRes.Coins[0].Amount is of pattern {amount}{denom} Ex: 438926033423uxyz
	// Hence, converting it to type coin to extract amount
	coin, err := sdktypes.ParseCoinNormalized(denomSupplyRes.Coins[0].Amount)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot convert amount to coin",
			http.StatusBadRequest,
		)
	}
	supply := coin.Amount.ToDec()

	//-----------------------------------------
	// get inflation

	inflation, err := getMintInflation(ctx, sdkClient, chain.ChainName)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve inflation from sdk-service",
			http.StatusBadRequest,
		)
	}
	// only 25% of the newly minted tokens are distributed as staking rewards for osmosis
	if strings.ToLower(chain.ChainName) == osmosisChainName {
		inflation = inflation.QuoInt64(4)
	}

	//-----------------------------------------
	// Compute APR

	return inflation.Quo(bondedTokens.Quo(supply)).MulInt64(100), nil
}

// apr=(1-budget rate)*(1-tax)*CurrentInflationAmount/Bonded tokens
func (app *App) getCrescentAPR(ctx context.Context, sdkClient sdkutilities.Service, chain cns.Chain, bondedTokens sdktypes.Dec) (sdktypes.Dec, error) {
	budgetRate, err := getBudgetRate(ctx, sdkClient, chain)
	if err != nil {
		return sdktypes.Dec{}, err
	}

	tax, err := getTax(ctx, sdkClient, chain)
	if err != nil {
		return sdktypes.Dec{}, err
	}

	currentInflationAmount, err := getCrescentCurrentInflation(ctx, sdkClient, chain)
	if err != nil {
		return sdktypes.Dec{}, err
	}

	oneDec := sdktypes.NewDec(1)
	return oneDec.Sub(tax).
		Mul(oneDec.Sub(budgetRate)).
		Mul(currentInflationAmount).
		Quo(bondedTokens).
		MulInt64(100), nil
}

func getBudgetRate(ctx context.Context, sdkClient sdkutilities.Service, chain cns.Chain) (sdktypes.Dec, error) {
	budgetParams, err := getBudgetParams(ctx, sdkClient, chain.ChainName)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve budget params from sdk-service",
			http.StatusBadRequest,
		)
	}

	const (
		ecosystemIncentiveBudget = "budget-ecosystem-incentive"
		devTeamBudget            = "budget-dev-team"
	)
	budgetRate := sdktypes.NewDec(0)
	for _, budget := range budgetParams.Params.Budgets {
		if budget.Name == ecosystemIncentiveBudget || budget.Name == devTeamBudget {
			rate, err := sdktypes.NewDecFromStr(budget.Rate)
			if err != nil {
				return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
					"cannot convert budget rate to Dec",
					http.StatusBadRequest,
				)
			}
			budgetRate = budgetRate.Add(rate)
		}
	}
	return budgetRate, nil
}

func getTax(ctx context.Context, sdkClient sdkutilities.Service, chain cns.Chain) (sdktypes.Dec, error) {
	distributionParams, err := getDistributionParams(ctx, sdkClient, chain.ChainName)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve distribution params from sdk-service",
			http.StatusBadRequest,
		)
	}

	tax, err := sdktypes.NewDecFromStr(distributionParams.Params.CommunityTax)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot convert tax to Dec",
			http.StatusBadRequest,
		)
	}
	return tax, nil
}

func getCrescentCurrentInflation(ctx context.Context, sdkClient sdkutilities.Service, chain cns.Chain) (sdktypes.Dec, error) {
	mintParams, err := getCrescentMintParams(ctx, sdkClient, chain.ChainName)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve mint params from sdk-service",
			http.StatusBadRequest,
		)
	}

	now := time.Now()
	for _, schedule := range mintParams.Params.InflationSchedules {
		if schedule.StartTime.Before(now) && schedule.EndTime.After(now) {
			currentInflationAmount, err := sdktypes.NewDecFromStr(schedule.Amount)
			if err != nil {
				return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
					"cannot convert amount to dec",
					http.StatusBadRequest,
				)
			}
			return currentInflationAmount, nil
		}
	}
	// Inflation not found in schedule, consider 0 ?
	return sdktypes.NewDec(0), nil
}
