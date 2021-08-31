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
	dbSaveResult := JsonSetData(db.Rejson, world.Name, ".", world)
	if dbSaveResult == "OK" {
		if verbose {
			log.Printf("Set world data %v", world)
		}
	} else {
		if verbose {
			log.Printf("Could not set world (%v). Error saving to DB. dbSaveResult : %v", world, dbSaveResult)
		}
	}
}

// Update world
func UpdateWorld (db Database, worldName string, path string, newValue interface{}) {
	world := GetUserData(db, worldName, ".")
	if (world != nil) {
		dbSaveResult := JsonSetData(db.Rejson, worldName, path, newValue)
		if dbSaveResult == "OK" {
			if verbose {
				log.Printf("Updated world data %v at path %v", worldName, path)
			}
		} else {
			if verbose {
				log.Printf("Could not update world (%v) at path %v. Error saving to DB. dbSaveResult : %v", worldName, path, dbSaveResult)
			}
		}
	} else {
		if verbose {
			log.Printf("UpdateWorld: %s attempted update but no world by the name exists", worldName)
		}
	}
}

// Get world and unmarshall into World
func GetWorld(db Database, world_name string) World {
	method := "GetWorld"
	dataJSON := JsonGetData(db.Rejson, world_name, ".")
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