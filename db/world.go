package db

import (
	"encoding/json"
	"log"

	structs "github.com/brct-james/guild-golems/structs"
)

type World struct {
	Name string `json:"world_name"`
	Regions []structs.Region `json:"regions"`
}

// Set world
func SetWorld (db Database, world World) {
	JsonSetData(db.Rejson, world.Name, world)
}

// Get world and unmarshall into World
func GetWorld(db Database, world_name string) World {
	method := "GetWorld"
	dataJSON := JsonGetData(db.Rejson, world_name)
	readData := World{}
	err := json.Unmarshal(dataJSON, &readData)
	if err != nil {
		log.Printf("%s: Failed to JSON Unmarshal. dataJSON: %s", method, dataJSON)
		return World{}
	}

	if verbose {
		log.Printf("%s: Unmarshalled data: %#v\n", method, readData)
	}
	return readData
}