// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"

	"github.com/brct-james/guild-golems/log"
)

// Defines the characteristics of Routes between locales
// TravelTime in seconds
type Route struct {
	Thing
	DangerLevel int `json:"danger_level" binding:"required"`
	TravelTime int `json:"travel_time" binding:"required"`
	Cost int `json:"cost" binding:"required"`
}

var Routes map[string]Route

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