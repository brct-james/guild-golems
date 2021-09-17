package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/brct-james/brct-game/auth"
	"github.com/brct-james/brct-game/filemngr"
	"github.com/brct-james/brct-game/log"
	"github.com/brct-james/brct-game/rdb"
	"github.com/brct-james/brct-game/schema"
)

// Configuration

var reloadWorldFromJSON bool = true
var refreshAuthSecret bool = true

var worldJSONPath string = "./v0_world.json"

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
var worldDatabase rdb.Database

// Main
func main() {
	log.Info.Printf("Brct-Game Rest API Server %s", apiVersion)
	log.Info.Printf("Connecting to Redis DB")
	
	userDatabase = rdb.NewDatabase(RedisAddr, dbMap["users"])
	worldDatabase = rdb.NewDatabase(RedisAddr, dbMap["world"])

	if reloadWorldFromJSON {
		log.Important.Printf("Flushing World Database")
		worldDatabase.Flush()
		log.Info.Println("Loading world.json -> DB")
		initializeWorldDB(worldDatabase)
	}

	if refreshAuthSecret {
		log.Important.Printf("(Re)Generating Auth Secret")
		auth.CreateOrUpdateAuthSecretInFile()
		log.Important.Printf("Flushing User Database")
		userDatabase.Flush()
	}

	log.Info.Println("Loading secrets from envfile")
	auth.LoadSecretsToEnv()

	// Begin serving
	// handleRequests()
}

// Load world file from json and save it to world database
func initializeWorldDB(wdb rdb.Database) {
	log.Info.Println("Unmarshaling world.json")
	var res schema.World
	err := json.Unmarshal(filemngr.ReadJSON(worldJSONPath), &res)
	if err != nil {
		log.Error.Fatalf("Could not unmarshal world.json: %v", err)
	}

	log.Info.Println("Saving json to DB")
	log.Debug.Printf("Json value:\n%v\n", res)
	err = wdb.SetJsonData("world",".",res)
	if err != nil {
		log.Error.Fatalf("Could not save world to DB: %v", err)
	}

	log.Debug.Printf("Getting world to ensure saved:\n")
	bytes, err := wdb.GetJsonData("world", ".")
	if err != nil {
		log.Error.Fatalf("Could not read world from DB: %v", err)
	}

	readData := schema.World{}
	err = json.Unmarshal(bytes, &readData)
	if err != nil {
		log.Error.Fatalf("Could not unmarshal world json from DB: %v", err)
	}
	success := fmt.Sprintf("%v", reflect.DeepEqual(readData, res))
	log.Test.Printf("DOES WORLD IN DB DEEPEQUAL WORLD FROM JSON? %s", log.TestOutput(success, "true"))
	if success != "true" {
		panic("FAILED TEST WHILE INITIALIZING WORLD DB, LOADED JSON NOT MATCH DATABASE")
	}
}

// func handleRequests() {
// 	//mux router
// 	mxr := mux.NewRouter().StrictSlash(true)
// 	mxr.Use(handlers.GenerateHandlerMiddlewareFunc(udb,wdb,worldName))
// 	mxr.HandleFunc("/", handlers.Homepage).Methods("GET")
// 	mxr.HandleFunc("/api", handlers.ApiSelection).Methods("GET")
// 	mxr.HandleFunc("/api/v0", handlers.V0Docs).Methods("GET")
// 	mxr.HandleFunc("/api/v0/status", handlers.V0Status).Methods("GET")
// 	mxr.HandleFunc("/api/v0/users", handlers.UsersSummary).Methods("GET")
// 	mxr.HandleFunc("/api/v0/users/{username}", handlers.UsernameInfo).Methods("GET")
// 	mxr.HandleFunc("/api/v0/users/{username}/claim", handlers.UsernameClaim).Methods("POST")

// 	// secure subrouter for account-specific routes
// 	secure := mxr.PathPrefix("/api/v0/my").Subrouter()
// 	secure.Use(auth.GenerateTokenValidationMiddlewareFunc(udb))
// 	secure.HandleFunc("/account", handlers.AccountInfo).Methods("GET")

// 	// Start listening
// 	log.Info.Println("Listening on :50242")
// 	log.Error.Fatal(http.ListenAndServe(":50242", mxr))
// }