// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"

	"github.com/brct-james/guild-golems/log"
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

var WorldInfo World

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