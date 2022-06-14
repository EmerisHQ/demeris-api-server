package usecase

import (
	"context"
	"encoding/json"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

type stakingPool struct {
	Pool struct {
		NotBondedTokens string `json:"not_bonded_tokens"`
		BondedTokens    string `json:"bonded_tokens"`
	} `json:"pool"`
}

func getStakingPool(ctx context.Context, sdk SDKServiceClient, chainName string) (stakingPool, error) {
	resp, err := sdk.StakingPool(ctx, &sdkutilities.StakingPoolPayload{
		ChainName: chainName,
	})
	if err != nil {
		return stakingPool{}, err
	}
	var sp stakingPool
	return sp, json.Unmarshal(resp.StakingPool, &sp)
}

type stakingParams struct {
	Params struct {
		UnbondingTime     int64  `json:"unbonding_time"`
		MaxValidators     int64  `json:"max_validators"`
		MaxEntries        int64  `json:"max_entries"`
		HistoricalEntries int64  `json:"historical_entries"`
		BondDenom         string `json:"bond_denom"`
	} `json:"params"`
}

func getStakingParams(ctx context.Context, sdk SDKServiceClient, chainName string) (stakingParams, error) {
	resp, err := sdk.StakingParams(ctx, &sdkutilities.StakingParamsPayload{
		ChainName: chainName,
	})
	if err != nil {
		return stakingParams{}, err
	}
	var sp stakingParams
	return sp, json.Unmarshal(resp.StakingParams, &sp)
}

func getMintInflation(ctx context.Context, sdk SDKServiceClient, chainName string) (sdktypes.Dec, error) {
	resp, err := sdk.MintInflation(ctx, &sdkutilities.MintInflationPayload{
		ChainName: chainName,
	})
	if err != nil {
		return sdktypes.Dec{}, err
	}

	var inflationData struct {
		Inflation string `json:"inflation"`
	}
	err = json.Unmarshal(resp.MintInflation, &inflationData)
	if err != nil {
		return sdktypes.Dec{}, err
	}
	return sdktypes.NewDecFromStr(inflationData.Inflation)
}

type budgetParams struct {
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

func getBudgetParams(ctx context.Context, sdk SDKServiceClient, chainName string) (budgetParams, error) {
	resp, err := sdk.BudgetParams(ctx, &sdkutilities.BudgetParamsPayload{
		ChainName: chainName,
	})
	if err != nil {
		return budgetParams{}, err
	}
	var bp budgetParams
	return bp, json.Unmarshal(resp.BudgetParams, &bp)
}

type distributionParams struct {
	Params struct {
		CommunityTax        string `json:"community_tax"`
		BaseProposerReward  string `json:"base_proposer_reward"`
		BonusProposerReward string `json:"bonus_proposer_reward"`
		WithdrawAddrEnabled bool   `json:"withdraw_addr_enabled"`
	} `json:"params"`
}

func getDistributionParams(ctx context.Context, sdk SDKServiceClient, chainName string) (distributionParams, error) {
	resp, err := sdk.DistributionParams(ctx,
		&sdkutilities.DistributionParamsPayload{ChainName: chainName})

	if err != nil {
		return distributionParams{}, err
	}
	var dp distributionParams
	return dp, json.Unmarshal(resp.DistributionParams, &dp)
}

type crescentMintParams struct {
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

func getCrescentMintParams(ctx context.Context, sdk SDKServiceClient, chainName string) (crescentMintParams, error) {
	resp, err := sdk.MintParams(ctx, &sdkutilities.MintParamsPayload{ChainName: chainName})
	if err != nil {
		return crescentMintParams{}, err
	}
	var mp crescentMintParams
	return mp, json.Unmarshal(resp.MintParams, &mp)
}
