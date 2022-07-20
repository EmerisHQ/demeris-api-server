package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/emerishq/demeris-api-server/api/database"
	"github.com/emerishq/demeris-api-server/lib/poclient"
	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/emerishq/demeris-backend-models/tracelistener"
)

// PriceOracle returns the price of a certain "symbol". A symbol is a string in
// the form "TICK1TICK2" where TICK1 and TICK2 are the ticker of the requested
// denoms pair. It's common to use "USDT" as the second ticker to get the price
// in USD.
type PriceOracle interface {
	GetPrice(ctx context.Context, symbol string) (poclient.Price, error)
}

type DenomTracer interface {
	DenomTrace(ctx context.Context, chainName, traceHash string) (tracelistener.IBCDenomTraceRow, error)
}

type Denomer interface {
	Denom(ctx context.Context, denomName string) (cns.Denom, error)
}

type DatabaseDenomer struct {
	db *database.Database
}

var _ Denomer = &DatabaseDenomer{}

func NewDatabaseDenomer(db *database.Database) *DatabaseDenomer {
	return &DatabaseDenomer{
		db: db,
	}
}

func (d *DatabaseDenomer) Denom(ctx context.Context, denomName string) (cns.Denom, error) {
	denoms, err := d.db.Denoms(ctx)
	if err != nil {
		return cns.Denom{}, fmt.Errorf("getting denoms: %w", err)
	}

	for _, denom := range denoms {
		if denom.Name == denomName {
			return denom, nil
		}
	}

	return cns.Denom{}, ErrDenomNotFound
}

var ErrDenomNotFound = errors.New("denom not found")

// PriceOracleDenomPricer implements the DenomPricer interface.
type PriceOracleDenomPricer struct {
	denomer     Denomer
	denomTracer DenomTracer
	priceOracle PriceOracle
}

var _ DenomPricer = &PriceOracleDenomPricer{}

func NewPriceOracleDenomPricer(denomer Denomer, denomTracer DenomTracer, priceOracle PriceOracle) *PriceOracleDenomPricer {
	return &PriceOracleDenomPricer{
		denomer:     denomer,
		denomTracer: denomTracer,
		priceOracle: priceOracle,
	}
}

func (p *PriceOracleDenomPricer) DenomPrice(ctx context.Context, chainName, denomName string) (sdktypes.Dec, error) {
	// resolve IBC denoms
	if denomName[:4] == "ibc/" {
		trace, err := p.denomTracer.DenomTrace(ctx, chainName, denomName[4:])
		if errors.Is(err, sql.ErrNoRows) {
			return sdktypes.Dec{}, fmt.Errorf("trace %s: %w", denomName, ErrIBCTraceNotFound)
		}
		if err != nil {
			return sdktypes.Dec{}, fmt.Errorf("resolving denom trace %s: %w", denomName, err)
		}
		denomName = trace.BaseDenom
		// TODO: should we also check if the trace is verified? how we do that?
	}

	// get denom
	denom, err := p.denomer.Denom(ctx, denomName)
	if err != nil {
		return sdktypes.Dec{}, fmt.Errorf("resolving denom ticker %s: %w", denomName, err)
	}

	// get price in USDT
	poPrice, err := p.priceOracle.GetPrice(ctx, denom.Ticker+"USDT")
	if err != nil {
		return sdktypes.Dec{}, fmt.Errorf("getting price for denom %s: %w", denom.Ticker, err)
	}

	// convert float into sdk.Dec
	poPrecision := math.Pow10(6)
	intPrice := int64(poPrice.Price * poPrecision)
	price := sdktypes.NewDec(intPrice).QuoInt64(int64(poPrecision))

	// get price for a single denom
	denomPrecision := math.Pow10(int(denom.Precision))
	price = price.QuoInt64(int64(denomPrecision))

	return price, nil
}

var ErrIBCTraceNotFound = errors.New("ibc trace not found")
