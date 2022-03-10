package account

import (
	"github.com/emerishq/demeris-backend-models/tracelistener"
)

type BalancesResponse struct {
	Balances []Balance `json:"balances"`
}
type Balance struct {
	Address   string  `json:"address,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified"`
	Amount    string  `json:"amount,omitempty"`
	OnChain   string  `json:"on_chain,omitempty"`
	Ibc       IbcInfo `json:"ibc,omitempty"`
}

type IbcInfo struct {
	Path string `json:"path,omitempty"`
	Hash string `json:"hash,omitempty"`
}

type StakingBalancesResponse struct {
	StakingBalances []StakingBalance `json:"staking_balances"`
}

type StakingBalance struct {
	ValidatorAddress string `json:"validator_address"`
	Amount           string `json:"amount"`
	ChainName        string `json:"chain_name"`
}

type UnbondingDelegationsResponse struct {
	UnbondingDelegations []UnbondingDelegation `json:"unbonding_delegations"`
}

type UnbondingDelegation struct {
	ValidatorAddress string                                   `json:"validator_address"`
	Entries          tracelistener.UnbondingDelegationEntries `json:"entries"`
	ChainName        string                                   `json:"chain_name"`
}

type NumbersResponse struct {
	Numbers []tracelistener.AuthRow `json:"numbers"`
}

type UserTicketsResponse struct {
	Tickets map[string][]string `json:"tickets"`
}

type DelegationDelegatorReward struct {
	ValidatorAddress string `json:"validator_address,omitempty"`
	Reward           string `json:"reward"`
}

type DelegatorRewardsResponse struct {
	Rewards []DelegationDelegatorReward `json:"rewards"`
	Total   string                      `json:"total"`
}
