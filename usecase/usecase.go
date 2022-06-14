package usecase

const (
	osmosisChainName  = "osmosis"
	crescentChainName = "crescent"
)

type App struct {
	db                DB
	sdkServiceClients SDKServiceClients
}

func NewApp(db DB, sdk SDKServiceClients) *App {
	return &App{
		db:                db,
		sdkServiceClients: sdk,
	}
}
