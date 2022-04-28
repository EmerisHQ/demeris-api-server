package apiutils

import (
	"context"
	"fmt"

	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/demeris-backend-models/tracelistener"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

// FetchAccountNumbers returns a tracelistener.AuthRow containing sequence
// and account numbers given a hex-encoded address.
func FetchAccountNumbers(chain cns.Chain, account string, sdkServiceClients sdkservice.SDKServiceClients) (tracelistener.AuthRow, error) {
	chainVersion := chain.MajorSDKVersion()
	chainName := chain.ChainName

	client, e := sdkServiceClients.GetSDKServiceClient(chainVersion)
	if e != nil {
		return tracelistener.AuthRow{}, fmt.Errorf(e.Error())
	}

	res, err := client.AccountNumbers(context.Background(), &sdkutilities.AccountNumbersPayload{
		ChainName:    chainName,
		Bech32Prefix: &chain.NodeInfo.Bech32Config.PrefixAccount,
		AddresHex:    &account,
	})
	if err != nil {
		return tracelistener.AuthRow{}, fmt.Errorf("cannot query account numbers, %w", err)
	}

	if res == nil {
		return tracelistener.AuthRow{}, nil
	}

	result := tracelistener.AuthRow{
		TracelistenerDatabaseRow: tracelistener.TracelistenerDatabaseRow{
			ChainName: chain.ChainName,
		},
		Address:        res.Bech32Address,
		SequenceNumber: uint64(res.SequenceNumber),
		AccountNumber:  uint64(res.AccountNumber),
	}

	return result, nil
}
