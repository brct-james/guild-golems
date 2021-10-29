// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"

	"github.com/brct-james/guild-golems/log"
)

// Defines generic resource type
type Resource struct {
	Thing
	CapacityPerUnit float64 `json:"capacity_per_unit" binding:"required"`
}

// Defines resource in an inventory, used in udb, not json/wdb
type InventoryResource struct {
	Resource
	Quantity int `json:"quantity" binding:"required"`
}

var Resources map[string]Resource

// Unmarshals resource from json byte array
func Resource_unmarshal_json(resource_json []byte) (Resource, error) {
	log.Debug.Println("Unmarshalling resource.json")
	var resource Resource
	err := json.Unmarshal(resource_json, &resource)
	if err != nil {
		return Resource{}, err
	}
	return resource, nil
}

// Unmarshals all resources from json byte array
func Resource_unmarshal_all_json(resource_json []byte) (map[string]Resource, error) {
	log.Debug.Println("Unmarshalling resource.json")
	nilRes := make(map[string]Resource)
	var resources map[string]Resource
	err := json.Unmarshal(resource_json, &resources)
	if err != nil {
		return nilRes, err
	}
	return resources, nil
}