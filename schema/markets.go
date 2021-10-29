// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
)

// Defines the characteristics of Markets
type Market struct {
	Thing
	Pricing map[string]PricingInfo `json:"pricing" binding:"required"`
	STock map[string]int `json:"stock" binding:"required"`
	Consumption map[string]int `json:"consumption" binding:"required"`
	Production map[string]int `json:"production" binding:"required"`
}

// Defines the characteristics of PricingInfo
type PricingInfo struct {
	Min int `json:"min" binding:"required"`
	Max int `json:"max" binding:"required"`
	Sensitivity int `json:"sensitivity" binding:"required"`
}

// Unmarshals market from json byte array
func Market_unmarshal_json(market_json []byte) (Market, error) {
	log.Debug.Println("Unmarshalling market.json")
	var market Market
	err := json.Unmarshal(market_json, &market)
	if err != nil {
		return Market{}, err
	}
	return market, nil
}

// Attempt to save market, returns error or nil
func Market_save_to_db(wdb rdb.Database, market Market) (error) {
	log.Debug.Printf("Saving market to DB")
	marketPath := fmt.Sprintf(".%s", market.Symbol)
	err := wdb.SetJsonData("markets", marketPath, market)
	return err
}

// Test: Get market from db and compare with json
func Test_market_initialized(wdb rdb.Database, market map[string]Market) {
	log.Debug.Printf("Comparing market db to expected value")
	market_data, getErr := Market_get_all_from_db(wdb)
	if getErr != nil {
		log.Error.Fatalf("Error encountered while testing market during wdb initialization: %v", getErr)
	}
	success_str := fmt.Sprintf("%v", reflect.DeepEqual(market_data, market))
	log.Test.Printf("%s DOES DB MARKET DEEPEQUAL JSON MARKET?", log.TestOutput(success_str, "true"))
	if success_str != "true" {
		log.Error.Fatalf("FAILED TEST WHILE INITIALIZING MARKET DB, LOADED JSON NOT MATCH DATABASE")
	}
}

// Get json from db based on path
func Market_get_json_from_db(wdb rdb.Database, path string) ([]byte, error) {
	log.Debug.Printf("Getting market json from db")
	bytes, err := wdb.GetJsonData("markets", path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Get the market specified by path from db
func Market_get_from_db(wdb rdb.Database, path string) (Market, error) {
	if strings.EqualFold(".", path) {
		log.Error.Printf("Calling market_get_from_db with . path, should use market_get_all_from_db instead!")
	}
	log.Debug.Printf("Getting market from db")
	bytes, getErr := Market_get_json_from_db(wdb, path)
	if getErr != nil {
		return Market{}, getErr
	}
	market, jsonErr := Market_unmarshal_json(bytes)
	if jsonErr != nil {
		return Market{}, jsonErr
	}
	return market, nil
}

// Attempt to save all markets, returns error or nil
func Market_save_all_to_db(wdb rdb.Database, markets map[string]Market) (error) {
	log.Debug.Printf("Saving all markets to DB")
	err := wdb.SetJsonData("markets", ".", markets)
	return err
}

// Unmarshals all markets from json byte array
func Market_unmarshal_all_json(market_json []byte) (map[string]Market, error) {
	log.Debug.Println("Unmarshalling market.json")
	nilRes := make(map[string]Market)
	var markets map[string]Market
	err := json.Unmarshal(market_json, &markets)
	if err != nil {
		return nilRes, err
	}
	return markets, nil
}

// Gets all markets from DB
func Market_get_all_from_db(wdb rdb.Database) (map[string]Market, error) {
	log.Debug.Printf("Getting all markets from db")
	nilRes := make(map[string]Market)
	bytes, getErr := Market_get_json_from_db(wdb, ".")
	if getErr != nil {
		log.Debug.Printf("GetError %v", getErr)
		return nilRes, getErr
	}
	markets, jsonErr := Market_unmarshal_all_json(bytes)
	if jsonErr != nil {
		log.Debug.Printf("JsonError %v", jsonErr)
		return nilRes, jsonErr
	}
	return markets, nil
}