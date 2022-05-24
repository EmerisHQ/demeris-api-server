package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/emerishq/demeris-api-server/lib/apierrors"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

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

func (app *App) StakingAPR(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error) {
	//-----------------------------------------
	// 1- get bonded tokens

	stakingPoolRes, err := app.sdkClient.StakingPool(ctx, &sdkutilities.StakingPoolPayload{
		ChainName: chain.ChainName,
	})
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve staking pool from sdk-service",
			http.StatusBadRequest,
		)
	}
	var stakingPoolData StakingPoolResponse
	err = json.Unmarshal(stakingPoolRes.StakingPool, &stakingPoolData)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot unmarshal staking pool",
			http.StatusBadRequest,
		)
	}

	bondedTokens, err := sdktypes.NewDecFromStr(stakingPoolData.Pool.BondedTokens)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			fmt.Sprintf("cannot convert bonded_tokens to sdktypes.Dec"),
			http.StatusBadRequest,
		)
	}

	// apr for crescent is calculated differently as it follows custom inflation schedules
	// apr=(1-budget rate)*(1-tax)*CurrentInflationAmount/Bonded tokens
	if strings.ToLower(chain.ChainName) == crescentChainName {
		return app.getCrescentAPR(ctx, chain, bondedTokens)
	}

	//-----------------------------------------
	// 2- get supply

	stakingParamsRes, err := app.sdkClient.StakingParams(ctx, &sdkutilities.StakingParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve staking params from sdk-service",
			http.StatusBadRequest,
		)
	}

	var stakingParamsData StakingParamsResponse
	err = json.Unmarshal(stakingParamsRes.StakingParams, &stakingParamsData)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot unmarshal staking params",
			http.StatusBadRequest,
		)
	}

	denomSupplyRes, err := app.sdkClient.SupplyDenom(ctx, &sdkutilities.SupplyDenomPayload{
		ChainName: chain.ChainName,
		Denom:     &stakingParamsData.Params.BondDenom,
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
				chain.ChainName, stakingParamsData.Params.BondDenom, len(denomSupplyRes.Coins)),
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

	inflationRes, err := app.sdkClient.MintInflation(ctx, &sdkutilities.MintInflationPayload{
		ChainName: chain.ChainName,
	})
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve inflation from sdk-service",
			http.StatusBadRequest,
		)
	}

	var inflationData struct {
		Inflation string `json:"inflation"`
	}
	err = json.Unmarshal(inflationRes.MintInflation, &inflationData)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot unmarshal inflation",
			http.StatusBadRequest,
		)
	}

	inflation, err := sdktypes.NewDecFromStr(inflationData.Inflation)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
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

	return inflation.Quo(bondedTokens.Quo(supply)).MulInt64(100), nil
}

// apr=(1-budget rate)*(1-tax)*CurrentInflationAmount/Bonded tokens
func (app *App) getCrescentAPR(ctx context.Context, chain cns.Chain, bondedTokens sdktypes.Dec) (sdktypes.Dec, error) {
	budgetRate, err := app.getBudgetRate(ctx, chain)
	if err != nil {
		return sdktypes.Dec{}, err
	}

	tax, err := app.getTax(ctx, chain)
	if err != nil {
		return sdktypes.Dec{}, err
	}

	currentInflationAmount, err := app.getCurrentInflationAmount(ctx, chain)
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

type BudgetParamsResponse struct {
	Params struct {
		EpochBlocks int64 `json:"epoch_blocks"`
		Budgets     []struct {
			Name               string `json:"name"`
			Rate               string `json:"rate"`
			SourceAddress      string `json:"source_address"`
			DestinationAddress string `json:"destination_address"`
			StartTime          string `json:"start_time"`
			EndTime            string `json:"end_time"`
		} `json:"budgets"`
	} `json:"params"`
}

func (app *App) getBudgetRate(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error) {
	budgetParamsResp, err := app.sdkClient.BudgetParams(ctx, &sdkutilities.BudgetParamsPayload{
		ChainName: chain.ChainName,
	})
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve budget params from sdk-service",
			http.StatusBadRequest,
		)
	}

	var budgetParamsData BudgetParamsResponse
	err = json.Unmarshal(budgetParamsResp.BudgetParams, &budgetParamsData)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot unmarshal budget params",
			http.StatusBadRequest,
		)
	}

	const (
		ecosystemIncentiveBudget = "budget-ecosystem-incentive"
		devTeamBudget            = "budget-dev-team"
	)
	budgetRate := sdktypes.NewDec(0)
	for _, budget := range budgetParamsData.Params.Budgets {
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

type DistributionParamsResponse struct {
	Params struct {
		CommunityTax        string `json:"community_tax"`
		BaseProposerReward  string `json:"base_proposer_reward"`
		BonusProposerReward string `json:"bonus_proposer_reward"`
		WithdrawAddrEnabled bool   `json:"withdraw_addr_enabled"`
	} `json:"params"`
}

func (app *App) getTax(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error) {
	distributionParamsResp, err := app.sdkClient.DistributionParams(ctx,
		&sdkutilities.DistributionParamsPayload{
			ChainName: chain.ChainName,
		})

	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve distribution params from sdk-service",
			http.StatusBadRequest,
		)
	}

	var distributionParamsData DistributionParamsResponse
	err = json.Unmarshal(distributionParamsResp.DistributionParams, &distributionParamsData)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot unmarshal distribution params",
			http.StatusBadRequest,
		)
	}

	tax, err := sdktypes.NewDecFromStr(distributionParamsData.Params.CommunityTax)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot convert tax to Dec",
			http.StatusBadRequest,
		)
	}
	return tax, nil
}

type CrecentMintParamsResponse struct {
	Params struct {
		MintDenom          string `json:"mint_denom"`
		BlockTimeThreshold string `json:"block_time_threshold"`
		InflationSchedules []struct {
			StartTime time.Time `json:"start_time"`
			EndTime   time.Time `json:"end_time"`
			Amount    string    `json:"amount"`
		} `json:"inflation_schedules"`
	} `json:"params"`
}

func (app *App) getCurrentInflationAmount(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error) {
	mintParamsResp, err := app.sdkClient.MintParams(ctx, &sdkutilities.MintParamsPayload{
		ChainName: chain.ChainName,
	})

	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot retrieve mint params from sdk-service",
			http.StatusBadRequest,
		)
	}

	var mintParamsData CrecentMintParamsResponse
	err = json.Unmarshal(mintParamsResp.MintParams, &mintParamsData)
	if err != nil {
		return sdktypes.Dec{}, apierrors.Wrap(err, "chains",
			"cannot unmarshal distribution params",
			http.StatusBadRequest,
		)
	}

	now := time.Now()
	for _, schedule := range mintParamsData.Params.InflationSchedules {
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
