// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"

	"github.com/brct-james/guild-golems/log"
)

// Defines the characteristics of Locales (e.g. cities, forests, etc.)
type Locale struct {
	Thing
	ResourceNodeSymbols []string `json:"resource_node_symbols" binding:"required"`
	RouteSymbols []string `json:"route_symbols" binding:"required"`
	MarketSymbols []string `json:"market_symbols" binding:"required"`
}

var Locales map[string]Locale

// Unmarshals locale from json byte array
func Locale_unmarshal_json(locale_json []byte) (Locale, error) {
	log.Debug.Println("Unmarshalling locale.json")
	var locale Locale
	err := json.Unmarshal(locale_json, &locale)
	if err != nil {
		return Locale{}, err
	}
	return locale, nil
}

// Unmarshals all locales from json byte array
func Locale_unmarshal_all_json(locale_json []byte) (map[string]Locale, error) {
	log.Debug.Println("Unmarshalling locale.json")
	nilRes := make(map[string]Locale)
	var locales map[string]Locale
	err := json.Unmarshal(locale_json, &locales)
	if err != nil {
		return nilRes, err
	}
	return locales, nil
}