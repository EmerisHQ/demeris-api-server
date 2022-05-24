package usecase

import (
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
)

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type App struct {
	sdkClient sdkutilities.Service
}

func NewApp(sdk sdkutilities.Service) *App {
	return &App{
		sdkClient: sdk,
	}
}
