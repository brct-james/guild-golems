// Package gamelogic provides functions for game logic
package gamelogic

import (
	"math"
	"time"

	"github.com/brct-james/guild-golems/gamevars"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/schema"
)

//schema.Market_get_from_db()
func calculateMarketTick(market schema.Market, elapsedTicks int) (schema.Market) {
	// Consumption
	for symbol, amount := range market.Consumption {
		market.Stock[symbol] = int(math.Max(0, float64(market.Stock[symbol] - (amount * elapsedTicks))))
	}
	// Production
	for symbol, amount := range market.Production {
		market.Stock[symbol] = market.Stock[symbol] + (amount * elapsedTicks)
	}
	return market
}

func CalculateAllMarketTicks(wdb rdb.Database) (map[string]schema.Market, error) {
	nilMkt := make(map[string]schema.Market)
	markets, mktGetErr := schema.Market_get_all_from_db(wdb)
	if mktGetErr != nil {
		return nilMkt, mktGetErr
	}
	// if enough time elapsed
	elapsedTime := time.Since(schema.LastMarketTick)
	elapsedSeconds := int64(elapsedTime / time.Second)
	elapsedTicks := int(elapsedSeconds / gamevars.Market_Consumption_Rate)
	if elapsedTicks < 1 {
		return markets, nil
	}
	for symbol, market := range markets {
		markets[symbol] = calculateMarketTick(market, elapsedTicks)
	}
	return markets, nil
}

func CalculateMarketPrice(mktPricing schema.PricingInfo, curQuant int) (int) {
// 	- - - - where min_price defines the horizontal asymptote (json_min_price - 1)
// - - - - where max_price_delta is the difference between min_price and max_price (json_max_price - min_price)
// - - - - where sensitivity defines the steepness of the curve (defines the vertical asymptote)
min_price := mktPricing.Min - 1
max_price_delta := mktPricing.Max - min_price
sensitivity := mktPricing.Sensitivity
quantity := curQuant
	return int((max_price_delta/(1+(quantity/sensitivity))) + min_price)+1
}