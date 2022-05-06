package poclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	potypes "github.com/allinbits/emeris-price-oracle/price-oracle/types"
	gocache "github.com/patrickmn/go-cache"
)

const DefaultCacheExpiration = 10 * time.Second
const DefaultPurgeTime = 15 * time.Second
const allPricesURL = "%s/prices"

type Price struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

type PricesResponse struct {
	Status  int                      `json:"status"`
	Data    potypes.AllPriceResponse `json:"data"`
	Message interface{}              `json:"message"`
}

type POClient struct {
	PriceOracleBaseURL string
	cache              *gocache.Cache
}

func NewPOClient(priceOracleBaseURL string) POClient {
	cache := gocache.New(DefaultCacheExpiration, DefaultPurgeTime)
	return POClient{
		PriceOracleBaseURL: priceOracleBaseURL,
		cache:              cache,
	}
}

// GetPrice returns price of token or fiat based on symbol
func (c POClient) GetPrice(symbol string) (Price, error) {
	// converting symbol to uppercase
	symbol = strings.ToUpper(symbol)

	res, found := c.cache.Get(symbol)
	if found {
		price, ok := res.(Price)
		if !ok {
			return Price{}, fmt.Errorf("cannot get price of %s from cache", symbol)
		}
		return price, nil
	}

	resp, err := http.Get(fmt.Sprintf(allPricesURL, c.PriceOracleBaseURL))
	if err != nil {
		return Price{}, err
	}

	if resp.StatusCode != 200 {
		return Price{}, fmt.Errorf("Got status code %d from price-oracle server", resp.StatusCode)
	}

	pricesRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Price{}, err
	}

	defer resp.Body.Close()

	var allPrices PricesResponse
	if err := json.Unmarshal(pricesRes, &allPrices); err != nil {
		return Price{}, err
	}

	// check all tokens to find given symbol
	for _, token := range allPrices.Data.Tokens {
		if token.Symbol == symbol {
			price := Price{
				Symbol: symbol,
				Price:  token.Price,
			}
			c.cache.Set(symbol, price, DefaultCacheExpiration)
			return price, nil
		}
	}

	// check all fiats to find given symbol
	for _, fiat := range allPrices.Data.Fiats {
		if fiat.Symbol == symbol {
			price := Price{
				Symbol: symbol,
				Price:  fiat.Price,
			}
			c.cache.Set(symbol, price, DefaultCacheExpiration)
			return price, nil
		}
	}

	return Price{}, fmt.Errorf("cannot get price for given symbol: %s", symbol)
}
