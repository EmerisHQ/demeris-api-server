package relayer

type RelayerStatusResponse struct {
	Running bool `json:"running"`
}

type RelayerBalance struct {
	Address       string `json:"address"`
	ChainName     string `json:"chain_name"`
	EnoughBalance bool   `json:"enough_balance"`
}

type RelayerBalances struct {
	Balances []RelayerBalance `json:"balances"`
}
