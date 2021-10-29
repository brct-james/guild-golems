// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"

	"github.com/brct-james/guild-golems/log"
)

type Region struct {
	Thing
	BorderRegionSymbols []string `json:"border_region_symbols" binding:"required"`
	LocaleSymbols []string `json:"locale_symbols" binding:"required"`
}

var Regions map[string]Region

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