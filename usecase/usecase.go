package usecase

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type App struct {
	sdkServiceClients SDKServiceClients
}

func NewApp(sdk SDKServiceClients) *App {
	return &App{
		sdkServiceClients: sdk,
	}
}
