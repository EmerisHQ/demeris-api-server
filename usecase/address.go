package usecase

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/getsentry/sentry-go"
)

// DeriveRawAddress returns the chain addresses from rawAddress for all enabled
// chains.
func (a *App) DeriveRawAddress(ctx context.Context, rawAddress string) ([]string, error) {
	defer sentry.StartSpan(ctx, "usecase.DeriveRawAddress").Finish()

	if rawAddress == "" {
		return nil, fmt.Errorf("raw address is empty")
	}
	bz, err := hex.DecodeString(rawAddress)
	if err != nil {
		return nil, fmt.Errorf("raw address is not in hex format: %w", err)
	}

	chains, err := a.db.Chains(ctx)
	if err != nil {
		return nil, err
	}
	addrs := make([]string, len(chains))
	for i, ch := range chains {
		// Get chain address bech 32 human readable part (aka prefix or tag)
		// FIXME(tb): MainPrefix or PrefixAccount or ?
		hrp := ch.NodeInfo.Bech32Config.MainPrefix
		addr, err := bech32.ConvertAndEncode(hrp, bz)
		if err != nil {
			return nil, err
		}
		addrs[i] = addr
	}
	return addrs, nil
}
