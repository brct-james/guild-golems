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

type Region struct {
	Thing
	BorderRegionSymbols []string `json:"border_region_symbols" binding:"required"`
	LocaleSymbols []string `json:"locale_symbols" binding:"required"`
}

// Unmarshals region from json byte array
func Region_unmarshal_json(region_json []byte) (Region, error) {
	log.Debug.Println("Unmarshalling region.json")
	var region Region
	err := json.Unmarshal(region_json, &region)
	if err != nil {
		return Region{}, err
	}
	return region, nil
}

// Attempt to save region, returns error or nil
func Region_save_to_db(wdb rdb.Database, region Region) (error) {
	log.Debug.Printf("Saving region to DB")
	regionPath := fmt.Sprintf(".%s", region.Symbol)
	err := wdb.SetJsonData("regions", regionPath, region)
	return err
}

// Test: Get region from db and compare with json
func Test_region_initialized(wdb rdb.Database, region map[string]Region) {
	log.Debug.Printf("Comparing region db to expected value")
	region_data, getErr := Region_get_all_from_db(wdb)
	if getErr != nil {
		log.Error.Fatalf("Error encountered while testing region during wdb initialization: %v", getErr)
	}
	success_str := fmt.Sprintf("%v", reflect.DeepEqual(region_data, region))
	log.Test.Printf("%s DOES DB REGION DEEPEQUAL JSON REGION?", log.TestOutput(success_str, "true"))
	if success_str != "true" {
		log.Error.Fatalf("FAILED TEST WHILE INITIALIZING REGION DB, LOADED JSON NOT MATCH DATABASE")
	}
}

// Get json from db based on path
func Region_get_json_from_db(wdb rdb.Database, path string) ([]byte, error) {
	log.Debug.Printf("Getting region json from db")
	bytes, err := wdb.GetJsonData("regions", path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Get the region specified by path from db
func Region_get_from_db(wdb rdb.Database, path string) (Region, error) {
	if strings.EqualFold(".", path) {
		log.Error.Printf("Calling region_get_from_db with . path, should use region_get_all_from_db instead!")
	}
	log.Debug.Printf("Getting region from db")
	bytes, getErr := Region_get_json_from_db(wdb, path)
	if getErr != nil {
		return Region{}, getErr
	}
	region, jsonErr := Region_unmarshal_json(bytes)
	if jsonErr != nil {
		return Region{}, jsonErr
	}
	return region, nil
}

// Attempt to save all regions, returns error or nil
func Region_save_all_to_db(wdb rdb.Database, regions map[string]Region) (error) {
	log.Debug.Printf("Saving all regions to DB")
	err := wdb.SetJsonData("regions", ".", regions)
	return err
}

// Unmarshals all regions from json byte array
func Region_unmarshal_all_json(region_json []byte) (map[string]Region, error) {
	log.Debug.Println("Unmarshalling region.json")
	nilRes := make(map[string]Region)
	var regions map[string]Region
	err := json.Unmarshal(region_json, &regions)
	if err != nil {
		return nilRes, err
	}
	return regions, nil
}

// Gets all regions from DB
func Region_get_all_from_db(wdb rdb.Database) (map[string]Region, error) {
	log.Debug.Printf("Getting all regions from db")
	nilRes := make(map[string]Region)
	bytes, getErr := Region_get_json_from_db(wdb, ".")
	if getErr != nil {
		log.Debug.Printf("GetError %v", getErr)
		return nilRes, getErr
	}
	regions, jsonErr := Region_unmarshal_all_json(bytes)
	if jsonErr != nil {
		log.Debug.Printf("JsonError %v", jsonErr)
		return nilRes, jsonErr
	}
	return regions, nil
}