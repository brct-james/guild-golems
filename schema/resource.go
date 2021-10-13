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

// Attempt to save resource, returns error or nil
func Resource_save_to_db(wdb rdb.Database, resource Resource) (error) {
	log.Debug.Printf("Saving resource to DB")
	resourcePath := fmt.Sprintf(".%s", resource.Symbol)
	err := wdb.SetJsonData("resources", resourcePath, resource)
	return err
}

// Test: Get resource from db and compare with json
func Test_resource_initialized(wdb rdb.Database, resource map[string]Resource) {
	log.Debug.Printf("Comparing resource db to expected value")
	resource_data, getErr := Resource_get_all_from_db(wdb)
	if getErr != nil {
		log.Error.Fatalf("Error encountered while testing resource during wdb initialization: %v", getErr)
	}
	success_str := fmt.Sprintf("%v", reflect.DeepEqual(resource_data, resource))
	log.Test.Printf("%s DOES DB RESOURCE DEEPEQUAL JSON RESOURCE?", log.TestOutput(success_str, "true"))
	if success_str != "true" {
		log.Error.Fatalf("FAILED TEST WHILE INITIALIZING RESOURCE DB, LOADED JSON NOT MATCH DATABASE")
	}
}

// Get json from db based on path
func Resource_get_json_from_db(wdb rdb.Database, path string) ([]byte, error) {
	log.Debug.Printf("Getting resource json from db")
	bytes, err := wdb.GetJsonData("resources", path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Get the resource specified by path from db
func Resource_get_from_db(wdb rdb.Database, path string) (Resource, error) {
	if strings.EqualFold(".", path) {
		log.Error.Printf("Calling resource_get_from_db with . path, should use resource_get_all_from_db instead!")
	}
	log.Debug.Printf("Getting resource from db")
	bytes, getErr := Resource_get_json_from_db(wdb, path)
	if getErr != nil {
		return Resource{}, getErr
	}
	resource, jsonErr := Resource_unmarshal_json(bytes)
	if jsonErr != nil {
		return Resource{}, jsonErr
	}
	return resource, nil
}

// Attempt to save all resources, returns error or nil
func Resource_save_all_to_db(wdb rdb.Database, resources map[string]Resource) (error) {
	log.Debug.Printf("Saving all resources to DB")
	err := wdb.SetJsonData("resources", ".", resources)
	return err
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

// Gets all resources from DB
func Resource_get_all_from_db(wdb rdb.Database) (map[string]Resource, error) {
	log.Debug.Printf("Getting all resources from db")
	nilRes := make(map[string]Resource)
	bytes, getErr := Resource_get_json_from_db(wdb, ".")
	if getErr != nil {
		log.Debug.Printf("GetError %v", getErr)
		return nilRes, getErr
	}
	resources, jsonErr := Resource_unmarshal_all_json(bytes)
	if jsonErr != nil {
		log.Debug.Printf("JsonError %v", jsonErr)
		return nilRes, jsonErr
	}
	return resources, nil
}