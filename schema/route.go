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

// Defines the characteristics of Routes between locales
// TravelTime in seconds
type Route struct {
	Thing
	DangerLevel int `json:"danger_level" binding:"required"`
	TravelTime int `json:"travel_time" binding:"required"`
	Cost int `json:"cost" binding:"required"`
}

// Unmarshals route from json byte array
func Route_unmarshal_json(route_json []byte) (Route, error) {
	log.Debug.Println("Unmarshalling route.json")
	var route Route
	err := json.Unmarshal(route_json, &route)
	if err != nil {
		return Route{}, err
	}
	return route, nil
}

// Attempt to save route, returns error or nil
func Route_save_to_db(wdb rdb.Database, route Route) (error) {
	log.Debug.Printf("Saving route to DB")
	routePath := fmt.Sprintf(".%s", route.Symbol)
	err := wdb.SetJsonData("routes", routePath, route)
	return err
}

// Test: Get route from db and compare with json
func Test_route_initialized(wdb rdb.Database, route map[string]Route) {
	log.Debug.Printf("Comparing route db to expected value")
	route_data, getErr := Route_get_all_from_db(wdb)
	if getErr != nil {
		log.Error.Fatalf("Error encountered while testing route during wdb initialization: %v", getErr)
	}
	success_str := fmt.Sprintf("%v", reflect.DeepEqual(route_data, route))
	log.Test.Printf("%s DOES DB ROUTE DEEPEQUAL JSON ROUTE?", log.TestOutput(success_str, "true"))
	if success_str != "true" {
		log.Error.Fatalf("FAILED TEST WHILE INITIALIZING ROUTE DB, LOADED JSON NOT MATCH DATABASE")
	}
}

// Get json from db based on path
func Route_get_json_from_db(wdb rdb.Database, path string) ([]byte, error) {
	log.Debug.Printf("Getting route json from db")
	bytes, err := wdb.GetJsonData("routes", path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Get the route specified by path from db
func Route_get_from_db(wdb rdb.Database, path string) (Route, error) {
	if strings.EqualFold(".", path) {
		log.Error.Printf("Calling route_get_from_db with . path, should use route_get_all_from_db instead!")
	}
	log.Debug.Printf("Getting route from db")
	bytes, getErr := Route_get_json_from_db(wdb, path)
	if getErr != nil {
		return Route{}, getErr
	}
	route, jsonErr := Route_unmarshal_json(bytes)
	if jsonErr != nil {
		return Route{}, jsonErr
	}
	return route, nil
}

// Attempt to save all routes, returns error or nil
func Route_save_all_to_db(wdb rdb.Database, routes map[string]Route) (error) {
	log.Debug.Printf("Saving all routes to DB")
	err := wdb.SetJsonData("routes", ".", routes)
	return err
}

// Unmarshals all routes from json byte array
func Route_unmarshal_all_json(route_json []byte) (map[string]Route, error) {
	log.Debug.Println("Unmarshalling route.json")
	nilRes := make(map[string]Route)
	var routes map[string]Route
	err := json.Unmarshal(route_json, &routes)
	if err != nil {
		return nilRes, err
	}
	return routes, nil
}

// Gets all routes from DB
func Route_get_all_from_db(wdb rdb.Database) (map[string]Route, error) {
	log.Debug.Printf("Getting all routes from db")
	nilRes := make(map[string]Route)
	bytes, getErr := Route_get_json_from_db(wdb, ".")
	if getErr != nil {
		log.Debug.Printf("GetError %v", getErr)
		return nilRes, getErr
	}
	routes, jsonErr := Route_unmarshal_all_json(bytes)
	if jsonErr != nil {
		log.Debug.Printf("JsonError %v", jsonErr)
		return nilRes, jsonErr
	}
	return routes, nil
}