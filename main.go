package main

import (
	"net/http"

	"github.com/brct-james/guild-golems/auth"
	"github.com/brct-james/guild-golems/filemngr"
	"github.com/brct-james/guild-golems/handlers"
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/schema"
	"github.com/gorilla/mux"
)

// Server Configuration

var reloadWorldFromJSON bool = true
var refreshAuthSecret bool = false
var flushUDB bool = true

var worldJSONPath string = "./static-files/json/v0_world.json"
var regionJSONPath string = "./static-files/json/v0_regions.json"
var localeJSONPath string = "./static-files/json/v0_locales.json"
var resourceJSONPath string = "./static-files/json/v0_resources.json"
var resourceNodeJSONPath string = "./static-files/json/v0_resource_nodes.json"
var routeJSONPath string = "./static-files/json/v0_routes.json"

// Game Configuration
// in user-metrics.go: activityThresholdInMinutes controls what users are considered 'active'

// Define relationship between string database name and redis db num
var dbMap = map[string]int{
	"users": 0,
	"world": 1,
}

// Global Vars

var apiVersion string = "v0.0.1"
var (
	ListenPort = ":50242"
	RedisAddr = "localhost:6380"
)

var userDatabase rdb.Database
var worldDatabase rdb.Database

// Main
func main() {
	log.Info.Printf("Guild-Golems Rest API Server %s", apiVersion)
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
	}

	if refreshAuthSecret || flushUDB {
		log.Important.Printf("Flushing User Database")
		userDatabase.Flush()
	}

	log.Info.Println("Loading secrets from envfile")
	auth.LoadSecretsToEnv()

	// Begin serving
	handle_requests()
}

// Load world file from json and save it to world database
func initializeWorldDB(wdb rdb.Database) {
	// --World--
	world, world_json_err := schema.World_unmarshal_json(filemngr.ReadJSON(worldJSONPath))
	if world_json_err != nil {
		log.Error.Fatalf("Could not unmarshal world json: %v", world_json_err)
	}
	world_save_err := schema.World_save_to_db(wdb, world)
	if world_save_err != nil {
		// Fail state, crash as world required
		log.Error.Fatalf("Failed saving world during wdb init, err: %v", world_save_err)
	}
	schema.Test_world_initialized(wdb, world)

	// --Regions--
	regions, region_json_err := schema.Region_unmarshal_all_json(filemngr.ReadJSON(regionJSONPath))
	if region_json_err != nil {
		log.Error.Fatalf("Could not unmarshal region json: %v", region_json_err)
	}
	region_save_err := schema.Region_save_all_to_db(wdb, regions)
	if region_save_err != nil {
		// Fail state, crash as region required
		log.Error.Fatalf("Failed saving region during wdb init, err: %v", region_save_err)
	}
	schema.Test_region_initialized(wdb, regions)

	// --Locales--
	locales, locale_json_err := schema.Locale_unmarshal_all_json(filemngr.ReadJSON(localeJSONPath))
	if locale_json_err != nil {
		log.Error.Fatalf("Could not unmarshal locale json: %v", locale_json_err)
	}
	locale_save_err := schema.Locale_save_all_to_db(wdb, locales)
	if locale_save_err != nil {
		// Fail state, crash as locale required
		log.Error.Fatalf("Failed saving locale during wdb init, err: %v", locale_save_err)
	}
	schema.Test_locale_initialized(wdb, locales)

	// --Routes--
	routes, route_json_err := schema.Route_unmarshal_all_json(filemngr.ReadJSON(routeJSONPath))
	if route_json_err != nil {
		log.Error.Fatalf("Could not unmarshal route json: %v", route_json_err)
	}
	route_save_err := schema.Route_save_all_to_db(wdb, routes)
	if route_save_err != nil {
		// Fail state, crash as route required
		log.Error.Fatalf("Failed saving route during wdb init, err: %v", route_save_err)
	}
	schema.Test_route_initialized(wdb, routes)

	// --Resources--
	resources, resource_json_err := schema.Resource_unmarshal_all_json(filemngr.ReadJSON(resourceJSONPath))
	if resource_json_err != nil {
		log.Error.Fatalf("Could not unmarshal resource json: %v", resource_json_err)
	}
	resource_save_err := schema.Resource_save_all_to_db(wdb, resources)
	if resource_save_err != nil {
		// Fail state, crash as resource required
		log.Error.Fatalf("Failed saving resource during wdb init, err: %v", resource_save_err)
	}
	schema.Test_resource_initialized(wdb, resources)

	// --Resource Nodes--
	resourceNodes, resourceNode_json_err := schema.ResourceNode_unmarshal_all_json(filemngr.ReadJSON(resourceNodeJSONPath))
	if resourceNode_json_err != nil {
		log.Error.Fatalf("Could not unmarshal resourcenode json: %v", resourceNode_json_err)
	}
	resourceNode_save_err := schema.ResourceNode_save_all_to_db(wdb, resourceNodes)
	if resourceNode_save_err != nil {
		// Fail state, crash as resourcenode required
		log.Error.Fatalf("Failed saving resourcenode during wdb init, err: %v", resourceNode_save_err)
	}
	schema.Test_resourcenode_initialized(wdb, resourceNodes)
}

func handle_requests() {
	//mux router
	mxr := mux.NewRouter().StrictSlash(true)
	mxr.Use(handlers.GenerateHandlerMiddlewareFunc(userDatabase,worldDatabase))
	mxr.HandleFunc("/", handlers.Homepage).Methods("GET")
	mxr.HandleFunc("/api", handlers.ApiSelection).Methods("GET")
	mxr.HandleFunc("/api/v0", handlers.V0Status).Methods("GET")
	mxr.HandleFunc("/api/v0/leaderboards", handlers.LeaderboardDescriptions).Methods("GET")
	mxr.HandleFunc("/api/v0/leaderboards/{board}", handlers.GetLeaderboards).Methods("GET")
	mxr.HandleFunc("/api/v0/users", handlers.UsersSummary).Methods("GET")
	mxr.HandleFunc("/api/v0/users/{username}", handlers.UsernameInfo).Methods("GET")
	mxr.HandleFunc("/api/v0/users/{username}/claim", handlers.UsernameClaim).Methods("POST")
	mxr.HandleFunc("/api/v0/locations", handlers.LocationsOverview).Methods("GET")

	// secure subrouter for account-specific routes
	secure := mxr.PathPrefix("/api/v0/my").Subrouter()
	secure.Use(auth.GenerateTokenValidationMiddlewareFunc(userDatabase))
	secure.HandleFunc("/account", handlers.AccountInfo).Methods("GET")
	secure.HandleFunc("/golems", handlers.GetGolems).Methods("GET")
	secure.HandleFunc("/golems/{archetype}", handlers.GetGolemsByArchetype).Methods("GET")
	secure.HandleFunc("/golem/{symbol}", handlers.GolemInfo).Methods("GET")
	secure.HandleFunc("/golem/{symbol}", handlers.ChangeGolemTask).Methods("PUT")
	secure.HandleFunc("/rituals", handlers.ListRituals).Methods("GET")
	secure.HandleFunc("/rituals/{ritual}", handlers.GetRitualInfo).Methods("GET")
	secure.HandleFunc("/rituals/summon-invoker", handlers.NewInvoker).Methods("POST")
	secure.HandleFunc("/rituals/summon-harvester", handlers.NewHarvester).Methods("POST")

	// Start listening
	log.Info.Printf("Listening on %s", ListenPort)
	log.Error.Fatal(http.ListenAndServe(ListenPort, mxr))
}