package usecase

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type App struct {
	sdkServiceClients SDKServiceClients
	osmosisClient     OsmosisClient
	crescentClient    CrescentClient
	denomPricer       DenomPricer
}

func NewApp(
	sdk SDKServiceClients,
	osmosisClient OsmosisClient,
	crescentClient CrescentClient,
	denomPricer DenomPricer,
) *App {
	return &App{
		sdkServiceClients: sdk,
		osmosisClient:     osmosisClient,
		crescentClient:    crescentClient,
		denomPricer:       denomPricer,
	}
}
