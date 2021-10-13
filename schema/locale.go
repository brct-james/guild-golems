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

// Defines the characteristics of Locales (e.g. cities, forests, etc.)
type Locale struct {
	Thing
	ResourceNodeSymbols []string `json:"resource_node_symbols" binding:"required"`
	RouteSymbols []string `json:"route_symbols" binding:"required"`
}

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

// Attempt to save locale, returns error or nil
func Locale_save_to_db(wdb rdb.Database, locale Locale) (error) {
	log.Debug.Printf("Saving locale to DB")
	localePath := fmt.Sprintf(".%s", locale.Symbol)
	err := wdb.SetJsonData("locales", localePath, locale)
	return err
}

// Test: Get locale from db and compare with json
func Test_locale_initialized(wdb rdb.Database, locale map[string]Locale) {
	log.Debug.Printf("Comparing locale db to expected value")
	locale_data, getErr := Locale_get_all_from_db(wdb)
	if getErr != nil {
		log.Error.Fatalf("Error encountered while testing locale during wdb initialization: %v", getErr)
	}
	success_str := fmt.Sprintf("%v", reflect.DeepEqual(locale_data, locale))
	log.Test.Printf("%s DOES DB LOCALE DEEPEQUAL JSON LOCALE?", log.TestOutput(success_str, "true"))
	if success_str != "true" {
		log.Error.Fatalf("FAILED TEST WHILE INITIALIZING LOCALE DB, LOADED JSON NOT MATCH DATABASE")
	}
}

// Get json from db based on path
func Locale_get_json_from_db(wdb rdb.Database, path string) ([]byte, error) {
	log.Debug.Printf("Getting locale json from db")
	bytes, err := wdb.GetJsonData("locales", path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Get the locale specified by path from db
func Locale_get_from_db(wdb rdb.Database, path string) (Locale, error) {
	if strings.EqualFold(".", path) {
		log.Error.Printf("Calling locale_get_from_db with . path, should use locale_get_all_from_db instead!")
	}
	log.Debug.Printf("Getting locale from db")
	bytes, getErr := Locale_get_json_from_db(wdb, path)
	if getErr != nil {
		return Locale{}, getErr
	}
	locale, jsonErr := Locale_unmarshal_json(bytes)
	if jsonErr != nil {
		return Locale{}, jsonErr
	}
	return locale, nil
}

// Attempt to save all locales, returns error or nil
func Locale_save_all_to_db(wdb rdb.Database, locales map[string]Locale) (error) {
	log.Debug.Printf("Saving all locales to DB")
	err := wdb.SetJsonData("locales", ".", locales)
	return err
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

// Gets all locales from DB
func Locale_get_all_from_db(wdb rdb.Database) (map[string]Locale, error) {
	log.Debug.Printf("Getting all locales from db")
	nilRes := make(map[string]Locale)
	bytes, getErr := Locale_get_json_from_db(wdb, ".")
	if getErr != nil {
		log.Debug.Printf("GetError %v", getErr)
		return nilRes, getErr
	}
	locales, jsonErr := Locale_unmarshal_all_json(bytes)
	if jsonErr != nil {
		log.Debug.Printf("JsonError %v", jsonErr)
		return nilRes, jsonErr
	}
	return locales, nil
}