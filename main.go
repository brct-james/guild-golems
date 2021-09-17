package main

import (
	"github.com/brct-james/brct-game/auth"
	"github.com/brct-james/brct-game/log"
	"github.com/brct-james/brct-game/rdb"
)

// Configuration

var resetDatabases bool = true
var refreshAuthSecret bool = true

// Define relationship between string database name and redis db num
var dbMap = map[string]int{
	"users": 0,
	"world": 1,
}

// Global Vars

var apiVersion string = "v0.0.1"
var (
	ListenAddr = "localhost:50235"
	RedisAddr = "localhost:6381"
)

var userDatabase rdb.Database

// Main
func main() {
	log.Info.Printf("Brct-Game Rest API Server %s", apiVersion)
	log.Info.Printf("Connecting to Redis DB")
	
	userDatabase = rdb.NewDatabase(RedisAddr, dbMap["users"])

	if resetDatabases {
		log.Important.Printf("Flushing All Databases")
		userDatabase.Flush()
		// TODO: Reinitialize world db
	}

	if refreshAuthSecret {
		log.Important.Printf("(Re)Generating Auth Secret")
		auth.CreateOrUpdateAuthSecretInFile()
		log.Important.Printf("Flushing User Database")
		userDatabase.Flush()
	}

	log.Info.Println("Loading secrets from envfile")
	auth.LoadSecretsToEnv()

	// Handle loading json to db
	log.Info.Println("Loading world json")
	// saveWorldJson(readJSON("./" + apiVersion + "_regions.json"), wdb)
	
	// Begin serving
	// handleRequests()
}