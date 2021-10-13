// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
)

type World struct {
	Thing
	RegionSymbols []string `json:"region_symbols" binding:"required"`
}

type WorldSummaryResponse struct {
	World World `json:"world" binding:"required"`
	Regions map[string]Region `json:"regions" binding:"required"`
	Locales map[string]Locale `json:"locales" binding:"required"`
	Routes map[string]Route `json:"routes" binding:"required"`
	Resources map[string]Resource `json:"resources" binding:"required"`
	ResourceNodes map[string]ResourceNode `json:"resource_nodes" binding:"required"`
}

// Unmarshals world from json byte array
func World_unmarshal_json(world_json []byte) (World, error) {
	log.Debug.Println("Unmarshalling world.json")
	var world World
	err := json.Unmarshal(world_json, &world)
	if err != nil {
		return World{}, err
	}
	return world, nil
}

// Attempt to save world, returns error or nil
func World_save_to_db(wdb rdb.Database, world World) (error) {
	log.Debug.Printf("Saving world to DB")
	err := wdb.SetJsonData("world", ".", world)
	return err
}

// Test: Get world from db and compare with json
func Test_world_initialized(wdb rdb.Database, world World) {
	log.Debug.Printf("Comparing world db to expected value")
	world_data, getErr := World_get_from_db(wdb, ".")
	if getErr != nil {
		log.Error.Fatalf("Error encountered while testing world during wdb initialization: %v", getErr)
	}
	success_str := fmt.Sprintf("%v", reflect.DeepEqual(world_data, world))
	log.Test.Printf("%s DOES DB WORLD DEEPEQUAL JSON WORLD?", log.TestOutput(success_str, "true"))
	if success_str != "true" {
		log.Error.Fatalf("FAILED TEST WHILE INITIALIZING WORLD DB, LOADED JSON NOT MATCH DATABASE")
	}
}

func World_get_json_from_db(wdb rdb.Database, path string) ([]byte, error) {
	log.Debug.Printf("Getting world json from db")
	bytes, err := wdb.GetJsonData("world", path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func World_get_from_db(wdb rdb.Database, path string) (World, error) {
	log.Debug.Printf("Getting world from db")
	bytes, getErr := World_get_json_from_db(wdb, path)
	if getErr != nil {
		return World{}, getErr
	}
	world, jsonErr := World_unmarshal_json(bytes)
	if jsonErr != nil {
		return World{}, jsonErr
	}
	return world, nil
}