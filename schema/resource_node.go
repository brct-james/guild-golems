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

// Defines harvestable resource node
type ResourceNode struct {
	Thing
	HarvestTime int `json:"harvest_time" binding:"required"`
	DropTables []DropTable `json:"drop_tables" binding:"required"`
}

// Defines droptables
type DropTable struct {
	ResourceSymbol string `json:"resource_symbol" binding:"required"`
	Rarity float64 `json:"rarity" binding:"required"`
	HarvestAmount int `json:"harvest_amount" binding:"required"`
}

// Unmarshals resourcenode from json byte array
func ResourceNode_unmarshal_json(resourcenode_json []byte) (ResourceNode, error) {
	log.Debug.Println("Unmarshalling resourcenode.json")
	var resourcenode ResourceNode
	err := json.Unmarshal(resourcenode_json, &resourcenode)
	if err != nil {
		return ResourceNode{}, err
	}
	return resourcenode, nil
}

// Attempt to save resourcenode, returns error or nil
func ResourceNode_save_to_db(wdb rdb.Database, resourcenode ResourceNode) (error) {
	log.Debug.Printf("Saving resourcenode to DB")
	resourcenodePath := fmt.Sprintf(".%s", resourcenode.Symbol)
	err := wdb.SetJsonData("resourcenodes", resourcenodePath, resourcenode)
	return err
}

// Test: Get resourcenode from db and compare with json
func Test_resourcenode_initialized(wdb rdb.Database, resourcenode map[string]ResourceNode) {
	log.Debug.Printf("Comparing resourcenode db to expected value")
	resourcenode_data, getErr := ResourceNode_get_all_from_db(wdb)
	if getErr != nil {
		log.Error.Fatalf("Error encountered while testing resourcenode during wdb initialization: %v", getErr)
	}
	success_str := fmt.Sprintf("%v", reflect.DeepEqual(resourcenode_data, resourcenode))
	log.Test.Printf("%s DOES DB RESOURCENODE DEEPEQUAL JSON RESOURCENODE?", log.TestOutput(success_str, "true"))
	if success_str != "true" {
		log.Error.Fatalf("FAILED TEST WHILE INITIALIZING RESOURCENODE DB, LOADED JSON NOT MATCH DATABASE")
	}
}

// Get json from db based on path
func ResourceNode_get_json_from_db(wdb rdb.Database, path string) ([]byte, error) {
	log.Debug.Printf("Getting resourcenode json from db")
	bytes, err := wdb.GetJsonData("resourcenodes", path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Get the resourcenode specified by path from db
func ResourceNode_get_from_db(wdb rdb.Database, path string) (ResourceNode, error) {
	if strings.EqualFold(".", path) {
		log.Error.Printf("Calling resourcenode_get_from_db with . path, should use resourcenode_get_all_from_db instead!")
	}
	log.Debug.Printf("Getting resourcenode from db")
	bytes, getErr := ResourceNode_get_json_from_db(wdb, path)
	if getErr != nil {
		return ResourceNode{}, getErr
	}
	resourcenode, jsonErr := ResourceNode_unmarshal_json(bytes)
	if jsonErr != nil {
		return ResourceNode{}, jsonErr
	}
	return resourcenode, nil
}

// Attempt to save all resourcenodes, returns error or nil
func ResourceNode_save_all_to_db(wdb rdb.Database, resourcenodes map[string]ResourceNode) (error) {
	log.Debug.Printf("Saving all resourcenodes to DB")
	err := wdb.SetJsonData("resourcenodes", ".", resourcenodes)
	return err
}

// Unmarshals all resourcenodes from json byte array
func ResourceNode_unmarshal_all_json(resourcenode_json []byte) (map[string]ResourceNode, error) {
	log.Debug.Println("Unmarshalling resourcenode.json")
	nilRes := make(map[string]ResourceNode)
	var resourcenodes map[string]ResourceNode
	err := json.Unmarshal(resourcenode_json, &resourcenodes)
	if err != nil {
		return nilRes, err
	}
	return resourcenodes, nil
}

// Gets all resourcenodes from DB
func ResourceNode_get_all_from_db(wdb rdb.Database) (map[string]ResourceNode, error) {
	log.Debug.Printf("Getting all resourcenodes from db")
	nilRes := make(map[string]ResourceNode)
	bytes, getErr := ResourceNode_get_json_from_db(wdb, ".")
	if getErr != nil {
		log.Debug.Printf("GetError %v", getErr)
		return nilRes, getErr
	}
	resourcenodes, jsonErr := ResourceNode_unmarshal_all_json(bytes)
	if jsonErr != nil {
		log.Debug.Printf("JsonError %v", jsonErr)
		return nilRes, jsonErr
	}
	return resourcenodes, nil
}