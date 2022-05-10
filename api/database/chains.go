package database

import (
	"database/sql"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	"github.com/lib/pq"
)

type ChainWithStatus struct {
	ID                  uint64                  `db:"id" json:"-"`
	Enabled             bool                    `db:"enabled" json:"enabled"`
	ChainName           string                  `db:"chain_name" json:"chain_name"`
	Logo                string                  `db:"logo" json:"logo"`
	DisplayName         string                  `db:"display_name" json:"display_name"`
	PrimaryChannel      cns.DbStringMap         `db:"primary_channel" json:"primary_channel"`
	Denoms              cns.DenomList           `db:"denoms" json:"denoms"`
	DemerisAddresses    pq.StringArray          `db:"demeris_addresses" json:"demeris_addresses"`
	GenesisHash         string                  `db:"genesis_hash" json:"genesis_hash"`
	NodeInfo            cns.NodeInfo            `db:"node_info" json:"node_info"`
	ValidBlockThresh    cns.Threshold           `db:"valid_block_thresh" json:"valid_block_thresh" swaggertype:"primitive,integer"`
	DerivationPath      string                  `db:"derivation_path" json:"derivation_path"`
	SupportedWallets    pq.StringArray          `db:"supported_wallets" json:"supported_wallets"`
	BlockExplorer       string                  `db:"block_explorer" json:"block_explorer"`
	PublicNodeEndpoints cns.PublicNodeEndpoints `db:"public_node_endpoints" json:"public_node_endpoints,omitempty"`
	CosmosSDKVersion    string                  `db:"cosmos_sdk_version" json:"cosmos_sdk_version,omitempty"`
	Online              bool                    `db:"online" json:"online" `
}

func (d *Database) Chain(name string) (cns.Chain, error) {
	var c cns.Chain

	n, err := d.dbi.DB.PrepareNamed(`
	SELECT
		id,
		enabled,
		chain_name,
		logo,
		display_name,
		primary_channel,
		denoms,
		demeris_addresses,
		genesis_hash,
		node_info,
		valid_block_thresh,
		derivation_path,
		supported_wallets,
		block_explorer,
		public_node_endpoints,
		cosmos_sdk_version
	FROM cns.chains
	WHERE chain_name=:name AND enabled=TRUE LIMIT 1
`)
	if err != nil {
		return cns.Chain{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}

func (d *Database) ChainExists(name string) (bool, error) {
	var exists bool
	query := `SELECT exists (
			SELECT
			id,
			enabled,
			chain_name,
			logo,
			display_name,
			primary_channel,
			denoms,
			demeris_addresses,
			genesis_hash,
			node_info,
			valid_block_thresh,
			derivation_path,
			supported_wallets,
			block_explorer,
			public_node_endpoints,
			cosmos_sdk_version
		FROM cns.chains
		WHERE chain_name=($1) AND enabled=TRUE LIMIT 1)`

	err := d.dbi.DB.QueryRow(query, name).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}

	return exists, err
}

func (d *Database) ChainFromChainID(chainID string) (cns.Chain, error) {
	var c cns.Chain

	n, err := d.dbi.DB.PrepareNamed(`
	SELECT
		id,
		enabled,
		chain_name,
		logo,
		display_name,
		primary_channel,
		denoms,
		demeris_addresses,
		genesis_hash,
		node_info,
		valid_block_thresh,
		derivation_path,
		supported_wallets,
		block_explorer,
		public_node_endpoints,
		cosmos_sdk_version
	FROM cns.chains
	WHERE node_info->>'chain_id'=:chainID AND enabled=TRUE LIMIT 1;
`)
	if err != nil {
		return cns.Chain{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Get(&c, map[string]interface{}{
		"chainID": chainID,
	})
}

func (d *Database) ChainLastBlock(name string) (tracelistener.BlockTimeRow, error) {
	var c tracelistener.BlockTimeRow

	n, err := d.dbi.DB.PrepareNamed(`
	SELECT
		id,
		chain_name,
		height,
		delete_height,
		block_time
	FROM tracelistener.blocktime 
	WHERE 
		chain_name=:name 
	AND 
		chain_name IN 
			(SELECT chain_name FROM cns.chains WHERE enabled=TRUE)
	`)
	if err != nil {
		return tracelistener.BlockTimeRow{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}

func (d *Database) Chains() ([]cns.Chain, error) {
	var c []cns.Chain
	return c, d.dbi.Exec(`
	SELECT
		id,
		enabled,
		chain_name,
		logo,
		display_name,
		primary_channel,
		denoms,
		demeris_addresses,
		genesis_hash,
		node_info,
		valid_block_thresh,
		derivation_path,
		supported_wallets,
		block_explorer,
		public_node_endpoints,
		cosmos_sdk_version
	FROM cns.chains where enabled=TRUE
	`, nil, &c)
}

func (d *Database) VerifiedDenoms() (map[string]cns.DenomList, error) {
	var c []cns.Chain
	if err := d.dbi.Exec("select chain_name, denoms from cns.chains where enabled=TRUE", nil, &c); err != nil {
		return nil, err
	}

	ret := make(map[string]cns.DenomList)

	for _, cc := range c {
		ret[cc.ChainName] = cc.VerifiedTokens()
	}

	return ret, nil
}

func (d *Database) SimpleChains() ([]cns.Chain, error) {
	var c []cns.Chain
	return c, d.dbi.Exec("select chain_name, display_name, logo from cns.chains where enabled=TRUE", nil, &c)
}

func (d *Database) ChainsWithStatus() ([]ChainWithStatus, error) {
	q := `
	SELECT
		c.enabled,c.chain_name,c.logo,c.display_name,c.primary_channel,c.denoms,c.demeris_addresses,
		c.genesis_hash,c.node_info,c.valid_block_thresh,c.derivation_path,c.supported_wallets,
		c.block_explorer,c.public_node_endpoints,c.cosmos_sdk_version,
		coalesce(
			parse_interval(c.valid_block_thresh) > current_timestamp() - b.block_time,
			false
		) online
	FROM cns.chains c
	LEFT JOIN tracelistener.blocktime b
	ON c.chain_name = b.chain_name
	WHERE c.enabled;
	`
	var rows []ChainWithStatus
	if err := d.dbi.Exec(q, nil, &rows); err != nil {
		return nil, err
	}

	return rows, nil
}

func (d *Database) ChainIDs() (map[string]string, error) {
	type it struct {
		ChainName string `db:"chain_name"`
		ChainID   string `db:"chain_id"`
	}

	c := map[string]string{}
	var cc []it
	err := d.dbi.Exec("select chain_name, node_info->>'chain_id' as chain_id from cns.chains where enabled=TRUE", nil, &cc)
	if err != nil {
		return nil, err
	}

	for _, ccc := range cc {
		c[ccc.ChainName] = ccc.ChainID
	}

	return c, nil
}

func (d *Database) PrimaryChannelCounterparty(chainName, counterparty string) (cns.ChannelQuery, error) {
	var c cns.ChannelQuery

	n, err := d.dbi.DB.PrepareNamed("select chain_name, mapping.* from cns.chains c, jsonb_each_text(primary_channel) mapping where key=:counterparty AND chain_name=:chain_name")
	if err != nil {
		return cns.ChannelQuery{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Get(&c, map[string]interface{}{
		"chain_name":   chainName,
		"counterparty": counterparty,
	})
}

func (d *Database) PrimaryChannels(chainName string) ([]cns.ChannelQuery, error) {
	var c []cns.ChannelQuery

	n, err := d.dbi.DB.PrepareNamed("select chain_name, mapping.* from cns.chains c, jsonb_each_text(primary_channel) mapping where chain_name=:chain_name")
	if err != nil {
		return nil, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Select(&c, map[string]interface{}{
		"chain_name": chainName,
	})
}

type ChainOnlineStatusRow struct {
	ChainName string `db:"chain_name"`
	Online    bool   `db:"online"`
}

func (d *Database) ChainsOnlineStatuses() ([]ChainOnlineStatusRow, error) {
	q := `
	SELECT
		c.chain_name,
		coalesce(
			parse_interval(c.valid_block_thresh) > current_timestamp() - b.block_time,
			false
		) online
	FROM cns.chains c
	LEFT JOIN tracelistener.blocktime b
	ON c.chain_name = b.chain_name
	WHERE c.enabled;
	`

	var rows []ChainOnlineStatusRow
	if err := d.dbi.DB.Select(&rows, q); err != nil {
		return nil, err
	}

	return rows, nil
}
