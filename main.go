package main

import (
	"github.com/brct-james/brct-game/log"
	"github.com/brct-james/brct-game/rdb"
)

// Configuration

var wipeDatabases bool = true
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

	if wipeDatabases {
		log.Important.Printf("Wiping Databases")
		userDatabase.Flush()
	}

	if refreshAuthSecret {
		log.Important.Printf("Refreshing Auth Secret")
		// Should this forcibly reset the user database? Almost certainly
	}
}