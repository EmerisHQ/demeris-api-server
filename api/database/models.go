package database

type ChannelConnectionMatchingDenoms []ChannelConnectionMatchingDenom

type ChannelConnectionMatchingDenom struct {
	ChainName             string `db:"chain_name"`
	ChannelId             string `db:"channel_id"`
	CounterpartyChain     string `db:"counterparty_chain"`
	CounterpartyChannelId string `db:"counterparty_channel_id"`
	BaseDenom             string `db:"base_denom"`
	Hash                  string `db:"hash"`
}
