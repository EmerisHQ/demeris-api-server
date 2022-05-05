package chains

import (
	"fmt"
	"strings"

	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/demeris-backend-models/tracelistener"
)

type ChainsResponse struct {
	Chains []database.ChainWithStatus `json:"chains"`
}
type SupportedChain struct {
	ChainName   string `json:"chain_name"`
	DisplayName string `json:"display_name"`
	Logo        string `json:"logo"`
}
type ChainResponse struct {
	Chain cns.Chain `json:"chain"`
}

type Bech32ConfigResponse struct {
	Bech32Config cns.Bech32Config `json:"bech32_config"`
}

type FeeResponse struct {
	Denoms cns.DenomList `json:"denoms"`
}

type FeeAddressResponse struct {
	FeeAddress []string `json:"fee_address"`
}
type FeeAddress struct {
	ChainName  string   `json:"chain_name"`
	FeeAddress []string `json:"fee_address"`
}
type FeeAddressesResponse struct {
	FeeAddresses []FeeAddress `json:"fee_addresses"`
}

type FeeTokenResponse struct {
	FeeTokens []cns.Denom `json:"fee_tokens"`
}

type PrimaryChannel struct {
	Counterparty string `json:"counterparty"`
	ChannelName  string `json:"channel_name"`
}

type PrimaryChannelResponse struct {
	Channel PrimaryChannel `json:"primary_channel"`
}
type PrimaryChannelsResponse struct {
	Channels []PrimaryChannel `json:"primary_channels"`
}

type Trace struct {
	Channel          string `json:"channel,omitempty"`
	Port             string `json:"port,omitempty"`
	ClientId         string `json:"client_id,omitempty"`
	ChainName        string `json:"chain_name,omitempty"`
	CounterpartyName string `json:"counterparty_name,omitempty"`
}

// IBCDenomHash represents the hash of an IBC denom. Its string representation
// follow the conventional format of the uppercased hash prefixed by "ibc/", e.g.:
//   ibc/ABC123XYZ
type IBCDenomHash string

func (d IBCDenomHash) Hash() string {
	return strings.ToUpper(string(d))
}

func (d IBCDenomHash) String() string {
	return fmt.Sprintf("ibc/%s", d.Hash())
}

func (d IBCDenomHash) MarshalJSON() ([]byte, error) {
	return []byte("\"" + d.String() + "\""), nil
}

type VerifiedTrace struct {
	// IbcDenom is the identifier of this denom in the form of "ibc/<hash>",
	// where <hash> is uppercased.
	IbcDenom IBCDenomHash `json:"ibc_denom,omitempty"`

	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified"`
	Path      string  `json:"path,omitempty"`
	Trace     []Trace `json:"trace,omitempty"`
	Cause     string  `json:"cause,omitempty"`
}

type VerifiedTraceResponse struct {
	VerifiedTrace VerifiedTrace `json:"verify_trace"`
}

type StatusResponse struct {
	Online bool `json:"online"`
}

type NumbersResponse struct {
	Numbers tracelistener.AuthRow `json:"numbers"`
}

type ValidatorsResponse struct {
	Validators []*Validator `json:"validators"`
}

type Validator struct {
	tracelistener.ValidatorRow
	Avatar string `json:"avatar,omitempty"`
}

//nolint //deadcode but used in swagger generation
type InflationResponse struct {
	Inflation string `json:"inflation"`
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

type StakingPoolResponse struct {
	Pool struct {
		NotBondedTokens string `json:"not_bonded_tokens"`
		BondedTokens    string `json:"bonded_tokens"`
	} `json:"pool"`
}

// nolint :ditto
type ParamsResponse struct {
	Params struct {
		MintDenom           string `json:"mint_denom"`
		InflationRateChange string `json:"inflation_rate_change"`
		InflationMax        string `json:"inflation_max"`
		InflationMin        string `json:"inflation_min"`
		GoalBonded          string `json:"goal_bonded"`
		BlocksPerYear       string `json:"blocks_per_year"`
	} `json:"params"`
}

// nolint :ditto
type AnnualProvisionsResponse struct {
	AnnualProvisions string `json:"annual_provisions"`
}

type Coin struct {
	Denom  string `json:"denom,omitempty"`
	Amount string `json:"amount,omitempty"`
}

type SupplyResponse struct {
	Supply     []Coin     `json:"supply,omitempty"`
	Pagination Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	NextKey string `json:"next_key,omitempty"`
	Total   string `json:"total,omitempty"`
}

type APRResponse struct {
	APR float64 `json:"apr,omitempty"`
}

type ChainStatus struct {
	Online bool `json:"online"`
}

type ChainsStatusesResponse struct {
	Chains map[string]ChainStatus `json:"chains"`
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

type DistributionParamsResponse struct {
	Params struct {
		CommunityTax        string `json:"community_tax"`
		BaseProposerReward  string `json:"base_proposer_reward"`
		BonusProposerReward string `json:"bonus_proposer_reward"`
		WithdrawAddrEnabled bool   `json:"withdraw_addr_enabled"`
	} `json:"params"`
}

type CrecentMintParamsResponse struct {
	Params struct {
		MintDenom          string `json:"mint_denom"`
		BlockTimeThreshold string `json:"block_time_threshold"`
		InflationSchedules []struct {
			StartTime string `json:"block_time_threshold"`
			EndTime   string `json:"end_time"`
			Amount    string `json:"amount"`
		}
	} `json:"params"`
}

func NewChainsStatusesResponse(sz int) ChainsStatusesResponse {
	return ChainsStatusesResponse{
		Chains: make(map[string]ChainStatus, sz),
	}
}
