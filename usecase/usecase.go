package usecase

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type App struct {
	sdkClient SDKClient
}

func NewApp(sdk SDKClient) *App {
	return &App{
		sdkClient: sdk,
	}
}
