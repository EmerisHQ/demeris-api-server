package usecase

import (
	"github.com/emerishq/demeris-api-server/sdkservice"
)

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type App struct {
	sdkServiceClients sdkservice.SDKServiceClients
}

func NewApp(sdk sdkservice.SDKServiceClients) *App {
	return &App{
		sdkServiceClients: sdk,
	}
}
