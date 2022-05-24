package usecase

import (
	"context"

	"github.com/emerishq/demeris-api-server/sdkservice"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/emeris-utils/exported/sdktypes"
)

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type IApp interface {
	StakingAPR(ctx context.Context, chain cns.Chain) (sdktypes.Dec, error)
}

type App struct {
	sdkServiceClients sdkservice.SDKServiceClients
}

func NewApp(sdk sdkservice.SDKServiceClients) IApp {
	return &App{
		sdkServiceClients: sdk,
	}
}
