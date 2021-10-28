// Package handlers provides handler functions for web routes
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/brct-james/guild-golems/auth"
	"github.com/brct-james/guild-golems/gamelogic"
	"github.com/brct-james/guild-golems/gamevars"
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/metrics"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/responses"
	"github.com/brct-james/guild-golems/schema"
	"github.com/gorilla/mux"
)

// HELPER FUNCTIONS

// ENUM for handler context
type HandlerResponseKey int
const (
	UserDBContext HandlerResponseKey = iota
	WorldDBContext
)

// Gets ResourceNode for target_node if in the locale specified by locale_path
func getTargetResNodeFromLocale(w http.ResponseWriter, r *http.Request, locale_path string, target_node string) (bool, schema.ResourceNode) {
	// Get wdb
	wdbSuccess, wdb := GetWdbFromCtx(w, r)
	if !wdbSuccess {
		log.Debug.Printf("Could not get wdb from ctx")
		return false, schema.ResourceNode{} // Fail state, could not get wdb, handled by func - simply return
	}
	// Success state, got wdb
	// Get locale data from db
	cur_locale, localeErr := schema.Locale_get_from_db(wdb, locale_path)
	if localeErr != nil {
		log.Error.Printf("Could not get locale %s from db: %v", locale_path, localeErr)
		responses.SendRes(w, responses.WDB_Get_Failure, nil, "locale corresponding to specified golem's location could not be gotten")
		return false, schema.ResourceNode{}
	}
	// Find resNodes for relevant locale
	if len(cur_locale.ResourceNodeSymbols) < 1 {
		// Fail case, no resNodes found
		log.Error.Printf("Golem has no available resource nodes! locale_path: %s | target_node: %s | Locale: %v", locale_path, target_node, cur_locale)
		responses.SendRes(w, responses.No_Resource_Nodes_At_Location, nil, "")
		return false, schema.ResourceNode{}
	}
	foundTargetNode, res_node := getResNodeDataIfTargetNodeInCurLocaleNodes(w, r, wdb, cur_locale, target_node)
	return foundTargetNode, res_node
}

// Gets route for target_route if in the locale specified by locale_path
func getTargetRouteFromLocale(w http.ResponseWriter, r *http.Request, locale_path string, target_route string) (bool, schema.Route) {
	// Get wdb
	wdbSuccess, wdb := GetWdbFromCtx(w, r)
	if !wdbSuccess {
		log.Debug.Printf("Could not get wdb from ctx")
		return false, schema.Route{} // Fail state, could not get wdb, handled by func - simply return
	}
	// Success state, got wdb
	// Get locale data from db
	cur_locale, localeErr := schema.Locale_get_from_db(wdb, locale_path)
	if localeErr != nil {
		log.Error.Printf("Could not get locale %s from db: %v", locale_path, localeErr)
		responses.SendRes(w, responses.WDB_Get_Failure, nil, "locale corresponding to specified golem's location could not be gotten")
		return false, schema.Route{}
	}
	// Find routes for relevant locale
	if len(cur_locale.RouteSymbols) < 1 {
		// Fail case, no routes found
		log.Error.Printf("Golem has no available routes! locale_path: %s | target_route: %s | Locale: %v", locale_path, target_route, cur_locale)
		responses.SendRes(w, responses.No_Available_Routes, nil, "")
		return false, schema.Route{}
	}
	foundTargetRoute, route := getRouteDataIfTargetRouteInCurLocaleRoutes(w, r, wdb, cur_locale, target_route)
	return foundTargetRoute, route
}

// Gets ResourceNode for target_node if in cur_locale's defined ResourceNodes
func getResNodeDataIfTargetNodeInCurLocaleNodes(w http.ResponseWriter, r *http.Request, wdb rdb.Database, cur_locale schema.Locale, target_node string) (bool, schema.ResourceNode) {
	// Now check for target_node in curLocale.ResourceNodes
	var target_node_symbol string
	nodeFound := false
	for _, nodeSymbol := range cur_locale.ResourceNodeSymbols {
		if strings.EqualFold(nodeSymbol, target_node) {
			// Success case, found target_node in curLocale.ResourceNodes
			log.Debug.Printf("Found target_node in curLocale.ResourceNodes. node.Symbol: %s, target_node.(string): %s", nodeSymbol, target_node)
			nodeFound = true
			target_node_symbol = nodeSymbol
			break
		}
	}
	if !nodeFound {
		// Fail case, target_node not in curLocale.ResourceNodes
		log.Debug.Printf("target_node not in curLocale.ResourceNodes")
		responses.SendRes(w, responses.Target_Resource_Node_Unavailable, nil, "")
		return false, schema.ResourceNode{}
	}

	// Get node data
	res_node_path := fmt.Sprintf("[\"%s\"]", target_node_symbol)
	cur_res_node, res_node_err := schema.ResourceNode_get_from_db(wdb, res_node_path)
	if res_node_err != nil {
		log.Error.Printf("Could not get resnode %s from db: %v", res_node_path, res_node_err)
		responses.SendRes(w, responses.WDB_Get_Failure, nil, "specified res node could not be gotten")
		return false, schema.ResourceNode{}
	}
	return true, cur_res_node
}

// Gets route for target_route if in cur_locale's defined routes
func getRouteDataIfTargetRouteInCurLocaleRoutes(w http.ResponseWriter, r *http.Request, wdb rdb.Database, cur_locale schema.Locale, target_route string) (bool, schema.Route) {
	// Now check for target_route in curLocale.Routes
	var target_route_symbol string
	routeFound := false
	for _, routeSymbol := range cur_locale.RouteSymbols {
		if strings.EqualFold(routeSymbol, target_route) {
			// Success case, found target_route in curLocale.Routes
			log.Debug.Printf("Found target_route in curLocal.Routes. route.Symbol: %s, target_route.(string): %s", routeSymbol, target_route)
			routeFound = true
			target_route_symbol = routeSymbol
			break
		}
	}
	if !routeFound {
		// Fail case, target_route not in curLocale.Routes
		log.Debug.Printf("target_route not in curLocale.Routes")
		responses.SendRes(w, responses.Target_Route_Unavailable, nil, "")
		return false, schema.Route{}
	}

	// Get route data
	route_path := fmt.Sprintf("[\"%s\"]", target_route_symbol)
	cur_route, route_err := schema.Route_get_from_db(wdb, route_path)
	if route_err != nil {
		log.Error.Printf("Could not get route %s from db: %v", route_path, route_err)
		responses.SendRes(w, responses.WDB_Get_Failure, nil, "specified route could not be gotten")
		return false, schema.Route{}
	}
	return true, cur_route
}

func executeGolemStatusChange(w http.ResponseWriter, r *http.Request, reqBody schema.GolemStatusUpdateBody, userData *schema.User, targetGolem *schema.Golem) {
	switch reqBody.NewStatus {
	case "idle":
		targetGolem.Status = "idle"
		targetGolem.StatusDetail = ""
		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	case "harvesting":
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		nodeSymbolInInstructions, target_node := stringKeyInMap("node_symbol", statusInstructions)
		if !nodeSymbolInInstructions {
			// Fail case
			log.Debug.Printf("'node_symbol' key required for 'harvesting' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'node_symbol' key required for 'harvesting' status")
			return
		}
		// Get ResourceNodes for golem locale
		locale_path := fmt.Sprintf("[\"%s\"]", targetGolem.LocationSymbol)
		gotNode, res_node := getTargetResNodeFromLocale(w, r, locale_path, target_node.(string))
		if !gotNode {
			return // Fail state, handled by func, return
		}
		// Success, found specified res node in locale and got data on it from wdb

		// Start harvesting
		targetGolem.Status = "harvesting"
		targetGolem.StatusDetail = res_node.Symbol
		// Save to DB
		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	case "traveling":
		// Check for all expected instructions
		// Convert Instructions to map with string keys
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		routeInInstructions, target_route := stringKeyInMap("route", statusInstructions)
		if !routeInInstructions {
			// Fail case
			log.Debug.Printf("'route' key required for 'traveling' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'route' key required for 'traveling' status")
			return
		}
		// Get routes for golem locale
		locale_path := fmt.Sprintf("[\"%s\"]", targetGolem.LocationSymbol)
		gotRoute, cur_route := getTargetRouteFromLocale(w, r, locale_path, target_route.(string))
		if !gotRoute {
			return // Fail state, handled by func, return
		}
		// Get destination from cur_route.Symbol
		log.Debug.Printf("cur_route.Symbol: %v", cur_route.Symbol)
		destinationSymbol := strings.Split(cur_route.Symbol, "|")[1]
		// Start travel
		// May set route danger as well later, to have a result calculated after travel completed
		targetGolem.TravelInfo.ArrivalTime = gamelogic.CalcualteArrivalTime(cur_route.TravelTime, targetGolem.Archetype).Unix()
		targetGolem.Status = "traveling"
		targetGolem.StatusDetail = destinationSymbol
		// Save to DB
		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	case "packing":
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		manifestInInstructions, tempManifest := stringKeyInMap("manifest", statusInstructions)
		if !manifestInInstructions {
			// Fail case
			log.Debug.Printf("'manifest' key required for 'packing/storing' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'manifest' key required for 'packing/storing' status")
			return
		}
		if len(tempManifest.(map[string]interface{})) < 1 {
			// Fail case
			responses.SendRes(w, responses.Blank_Manifest_Disallowed, nil, "")
			return
		}
		// Convert manifest to map[string]int from map[string]interface{}
		c, ok := tempManifest.(map[string]interface{})
		if !ok {
			// cant assert, handle error
			log.Error.Printf("cant assert tempManifest.(map[string]interface{}")
			responses.SendRes(w, responses.Generic_Failure, nil, "Cant assert tempManifest.(map[string]interface{}")
			return
		}
		manifest := make(map[string]int)
		for k,v := range c {
			manifest[k] = int(v.(float64))
		}

		// Get wdb
		wdbSuccess, wdb := GetWdbFromCtx(w, r)
		if !wdbSuccess {
			log.Debug.Printf("Could not get wdb from ctx")
			return // Fail state, could not get wdb, handled by func - simply return
		}
		// Success state, got wdb

		// Get resources in locale inventory at golem location
		
		gotInv, locInv := schema.GetInventoryByKey(targetGolem.LocationSymbol, userData.Inventories)
		if !gotInv {
			// Fail case - no items in inventory at location
			responses.SendRes(w, responses.No_Packable_Items, nil, "")
			return
		}

		// Validate and handle packing manifest
		for symbol, quantity := range manifest {
			// Check inv for each item in manifest
			contains, amountContained := schema.DoesInventoryContain(locInv, symbol, quantity)
			if !contains {
				// Fail case - invalid manifest, specified item not contained in sufficient quantity at location
				resMsg := fmt.Sprintf("Item: %s, Amount contained: %d", symbol, amountContained)
				responses.SendRes(w, responses.Invalid_Manifest, nil, resMsg)
				return
			}
			// Success case - update inventories with new contents
			userData.Inventories[targetGolem.LocationSymbol].Contents[symbol] = amountContained - quantity
			// delete symbol in contents if empty, then if contents empty delete entry in Inventories
			if amountContained - quantity == 0 {
				if len(userData.Inventories[targetGolem.LocationSymbol].Contents) == 1 {
					delete(userData.Inventories, targetGolem.LocationSymbol)
				}
				delete(userData.Inventories[targetGolem.LocationSymbol].Contents, symbol)
			} 
			// Check for inventory with specified symbol, if not exist already then make it exist
			if _, ok := userData.Inventories[targetGolem.Symbol]; !ok {
				userData.Inventories[targetGolem.Symbol] = schema.Inventory{LocationSymbol:targetGolem.Symbol,Contents:make(map[string]int)}
			}
			userData.Inventories[targetGolem.Symbol].Contents[symbol] = userData.Inventories[targetGolem.Symbol].Contents[symbol] + quantity
		}

		// Validate golem inventory can hold the amount of items specified in manifest before saving to db
		newCapacity := 0.0
		for symbol, quantity := range userData.Inventories[targetGolem.Symbol].Contents {
			res_path := fmt.Sprintf("[\"%s\"]", symbol)
			res, getResErr := schema.Resource_get_from_db(wdb, res_path)
			if getResErr != nil {
				// fail state, could not get res from db'
				resMsg := fmt.Sprintf("res_path: %s", res_path)
				responses.SendRes(w, responses.WDB_Get_Failure, nil, resMsg)
				return
			}
			// success state
			newCapacity += res.CapacityPerUnit * float64(quantity)
		}
		if newCapacity > targetGolem.Capacity {
			// fail state, newCapacity exceeds golem max
			resMsg := fmt.Sprintf("newCapacity %f exceeds golem max %f", newCapacity, targetGolem.Capacity)
			responses.SendRes(w, responses.Manifest_Overflow, nil, resMsg)
			return
		}
		
		// success state - save to db

		// TODO: make packing and storing take time (so these statuses don't finish instantly)
		targetGolem.Status = "idle"
		targetGolem.StatusDetail = ""

		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}

		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	case "storing":
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		manifestInInstructions, tempManifest := stringKeyInMap("manifest", statusInstructions)
		if !manifestInInstructions {
			// Fail case
			log.Debug.Printf("'manifest' key required for 'packing/storing' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'manifest' key required for 'packing/storing' status")
			return
		}
		if len(tempManifest.(map[string]interface{})) < 1 {
			// Fail case
			responses.SendRes(w, responses.Blank_Manifest_Disallowed, nil, "")
			return
		}
		// Convert manifest to map[string]int from map[string]interface{}
		c, ok := tempManifest.(map[string]interface{})
		if !ok {
			// cant assert, handle error
			log.Error.Printf("cant assert tempManifest.(map[string]interface{}")
			responses.SendRes(w, responses.Generic_Failure, nil, "Cant assert tempManifest.(map[string]interface{}")
			return
		}
		manifest := make(map[string]int)
		for k,v := range c {
			manifest[k] = int(v.(float64))
		}
		
		// Get resources in golem inventory
		gotInv, golInv := schema.GetInventoryByKey(targetGolem.Symbol, userData.Inventories)
		if !gotInv {
			// Fail case - no items in inventory
			responses.SendRes(w, responses.No_Storable_Items, nil, "")
			return
		}

		// Validate and handle storing manifest
		for symbol, quantity := range manifest {
			// Check inv for each item in manifest
			contains, amountContained := schema.DoesInventoryContain(golInv, symbol, quantity)
			if !contains {
				// Fail case - invalid manifest, specified item not contained in sufficient quantity
				resMsg := fmt.Sprintf("Item: %s, Amount contained: %d", symbol, amountContained)
				responses.SendRes(w, responses.Invalid_Manifest, nil, resMsg)
				return
			}
			// Success case - update inventories with new contents
			userData.Inventories[targetGolem.Symbol].Contents[symbol] = amountContained - quantity
			// delete symbol in contents if empty, then if contents empty delete entry in Inventories
			if amountContained - quantity == 0 {
				if len(userData.Inventories[targetGolem.Symbol].Contents) == 1 {
					delete(userData.Inventories, targetGolem.Symbol)
				}
				delete(userData.Inventories[targetGolem.Symbol].Contents, symbol)
			} 
			// Check for inventory with specified symbol, if not exist already then make it exist
			if _, ok := userData.Inventories[targetGolem.LocationSymbol]; !ok {
				userData.Inventories[targetGolem.LocationSymbol] = schema.Inventory{LocationSymbol:targetGolem.LocationSymbol,Contents:make(map[string]int)}
			}
			userData.Inventories[targetGolem.LocationSymbol].Contents[symbol] = userData.Inventories[targetGolem.LocationSymbol].Contents[symbol] + quantity
		}
		
		// success state - save to db

		// TODO: make packing and storing take time (so these statuses don't finish instantly)
		targetGolem.Status = "idle"
		targetGolem.StatusDetail = ""

		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}

		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	case "invoking":
		// Start invoking
		targetGolem.Status = "invoking"
		targetGolem.StatusDetail = ""
		// Save to DB
		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	default:
		// Error state, newStatus passed validation but not caught by switch statement
		//TODO: this
		responses.SendRes(w, responses.Generic_Failure, nil, "Unexpected Error state, newStatus passed validation but not caught by switch statement. Contact developer")
	}
}

// Attempt to get validation context
func GetValidationFromCtx(r *http.Request) (auth.ValidationPair, error) {
	log.Debug.Println("Recover validationpair from context")
	userInfo, ok := r.Context().Value(auth.ValidationContext).(auth.ValidationPair)
	if !ok {
		return auth.ValidationPair{}, errors.New("could not get ValidationPair")
	}
	// Any time a user hits a secure endpoint, track a call from their account
	metrics.TrackUserCall(userInfo.Username)
	return userInfo, nil
}

// Generates middleware func to pass databases to handlers using context
func GenerateHandlerMiddlewareFunc(udb rdb.Database, wdb rdb.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug.Println(log.Yellow("-- GenerateHandlerMiddlewareFunc --"))
			// Utilize context package to pass data to routes from the middleware
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserDBContext, udb)
			ctx = context.WithValue(ctx, WorldDBContext, wdb)
			r = r.WithContext(ctx)
			next.ServeHTTP(w,r)
			log.Debug.Println(log.Cyan("-- End GenerateHandlerMiddlewareFunc --"))
		})
	}
}

// Get User from Middleware and DB
// Returns: OK, userData, udb, userAuthPair
func secureGetUser(w http.ResponseWriter, r *http.Request) (bool, schema.User, rdb.Database, auth.ValidationPair) {
	// Get udb from context
	udb, udbErr := GetUdbFromCtx(r)
	if udbErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get UserDBContext in secureGetUser")
		responses.SendRes(w, responses.No_UDB_Context, nil, "in secureGetUser")
		return false, schema.User{}, rdb.Database{}, auth.ValidationPair{}
	}
	// Get userinfoContext from validation middleware
	userInfo, userInfoErr := GetValidationFromCtx(r)
	if userInfoErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get validationpair in secureGetUser")
		userInfoErrMsg := fmt.Sprintf("userInfo is nil, check auth validation context %v:\n%v", auth.ValidationContext, r.Context().Value(auth.ValidationContext))
		responses.SendRes(w, responses.No_AuthPair_Context, nil, userInfoErrMsg)
		return false, schema.User{}, rdb.Database{}, auth.ValidationPair{}
	}
	log.Debug.Printf("Validated with username: %s and token %s", userInfo.Username, userInfo.Token)
	// Check db for user
	thisUser, userFound, getUserErr := schema.GetUserFromDB(userInfo.Token, udb)
	if getUserErr != nil {
		// fail state
		getErrorMsg := fmt.Sprintf("in secureGetUser, could not get from DB for username: %s, error: %v", userInfo.Username, getUserErr)
		responses.SendRes(w, responses.UDB_Get_Failure, nil, getErrorMsg)
		return false, schema.User{}, rdb.Database{}, auth.ValidationPair{}
	}
	if !userFound {
		// fail state - user not found
		userNotFoundMsg := fmt.Sprintf("in secureGetUser, no user found in DB with username: %s", userInfo.Username)
		responses.SendRes(w, responses.User_Not_Found, nil, userNotFoundMsg)
		return false, schema.User{}, rdb.Database{}, auth.ValidationPair{}
	}

	// Get wdb
	wdbSuccess, wdb := GetWdbFromCtx(w, r)
	if !wdbSuccess {
		log.Debug.Printf("Could not get wdb from ctx")
		return false, schema.User{}, rdb.Database{}, auth.ValidationPair{} // Fail state, could not get wdb, handled by func - simply return
	}
	// Success state, got wdb

	// Success case
	thisUser, calcErr := gamelogic.CalculateUserUpdates(thisUser, wdb)
	if calcErr != nil {
		// Fail state could not calculate user updates
		resMsg := fmt.Sprintf("calcErr: %v", calcErr)
		responses.SendRes(w, responses.Generic_Failure, nil, resMsg)
		return false, thisUser, udb, userInfo
	}
	return true, thisUser, udb, userInfo
}

// Check user data for ritual in list of known rituals
func doesUserKnowRitual(userData schema.User, ritualKey string) (bool) {
	for _, ritual := range userData.KnownRituals {
		if strings.EqualFold(ritual, ritualKey) {
			// success state - user knows ritual
			return true
		}
	}
	// fail state - user doesnt know ritual
	return false
}

// Create new golem for user in database, if able, of particular archetype
func createNewGolemInDB(w http.ResponseWriter, r *http.Request, udb rdb.Database, userData schema.User, archetype string, locationSymbol string, ritualName string, startingStatus string, capacity float64) (bool) {
	knowsRitual := doesUserKnowRitual(userData, ritualName)
	if !knowsRitual {
		responses.SendRes(w, responses.Ritual_Not_Known, nil, "")
		return false
	}
	success, newManaValue := gamelogic.TryManaPurchase(w, userData.Mana, 600)
	if !success {
		return false // Failure states handled by TryManaPurchase, return false for failure
	}
	userData.Mana = newManaValue
	// // At this time not able to delete or lose golems so using len() on filtered list of golems is fine
	// var sameArchetypeGolemIds []int
	// for _, golem := range schema.FilterGolemListByArchetype(userData.Golems, archetype) {
	// 	golemId, err := strconv.Atoi(strings.Split(golem.Symbol, "-")[1])
	// 	if err != nil {
	// 		// Failure state - could not convert id from string to int
	// 		responses.SendRes(w, responses.Generic_Failure, nil, "Internal server error in createNewGolemInDB")
	// 		return false
	// 	}
	// 	sameArchetypeGolemIds = append(sameArchetypeGolemIds, golemId)
	// }
	// // todo: sort and use ids[-1]+1 for newGolemId
	var newGolemId int = len(schema.FilterGolemListByArchetype(userData.Golems, archetype))
	
	newGolemSymbol := fmt.Sprintf("%s-%d", schema.GolemArchetypes[archetype].Abbreviation, newGolemId)
	newGolem := schema.NewGolem(newGolemSymbol, archetype, locationSymbol, startingStatus, capacity)
	userData.Golems = append(userData.Golems, newGolem)
	saveUserErr := SaveUserToDB(udb, userData)
	if saveUserErr != nil {
		// fail state - could not save
		saveUserErrMsg := fmt.Sprintf("in createNewGolemInDB | Username: %v | SaveUserToDB failed, dbSaveResult: %v", userData.Username, saveUserErr)
		log.Debug.Println(saveUserErrMsg)
		responses.SendRes(w, responses.DB_Save_Failure, nil, saveUserErrMsg)
		return false
	}
	// Updated successfully
	log.Debug.Printf("Spawned new %s golem for username %s", archetype, userData.Username)
	responses.SendRes(w, responses.Generic_Success, newGolem, "")
	return true
}

// Remove trailing s if one exists
func trimTrailingS(input string) (string) {
	size := len(input)
	if size > 0 && input[size-1] == 's' {
		return input[:size-1]
	}
	return input
}

// Get body for statusUpdate requests
func getRequestBodyForGolemStatusUpdate(w http.ResponseWriter, r *http.Request) (bool, schema.GolemStatusUpdateBody) {
	var body schema.GolemStatusUpdateBody
	decoder := json.NewDecoder(r.Body)
	if decodeErr := decoder.Decode(&body); decodeErr != nil {
		// Fail case, could not decode
		responses.SendRes(w, responses.Bad_Request, nil, "Could not decode request body, is it present?")
		log.Debug.Printf("Error in getRequestBodyForGolemStatusUpdate: %v", decodeErr)
		return false, schema.GolemStatusUpdateBody{}
	}
	// Success case, decoded request
	return true, body
}

func stringKeyInMap(key string, dict map[string]interface{}) (bool, interface{}) {
	if val, ok := dict[key]; ok {
		// yes, key in map
		return true, val
	}
	// no, key not in map
	return false, nil
}

func GetUDBAndSaveUserToDB(w http.ResponseWriter, r *http.Request, userData schema.User) (bool) {
	udb, udbErr := GetUdbFromCtx(r)
	if udbErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get UserDBContext in GetUDBAndSaveUserToDB")
		responses.SendRes(w, responses.UDB_Update_Failed, nil, "")
		return false
	}
	saveUserErr := SaveUserToDB(udb, userData)
	if saveUserErr != nil {
		// fail state - could not save
		saveUserErrMsg := fmt.Sprintf("in GetUDBAndSaveUserToDB | Username: %v | SaveUserToDB failed, dbSaveResult: %v", userData.Username, saveUserErr)
		log.Debug.Println(saveUserErrMsg)
		responses.SendRes(w, responses.DB_Save_Failure, nil, saveUserErrMsg)
		return false
	}
	return true
}

func checkStatusChangeAllowedAndGetReqBody(w http.ResponseWriter, r *http.Request, currentStatus string, archetype string) (bool, schema.GolemStatusUpdateBody) {
	// Found golem, check that not in blocking status
	statusInfo, ok := schema.GolemStatuses[currentStatus]
	if !ok {
		// Fail case - golem status not in list of valid statuses
		resMsg1 := fmt.Sprintf("in ChangeGolemTask, golem status %s not in list of valid statuses", currentStatus)
		responses.SendRes(w, responses.Generic_Failure, nil, resMsg1)
		return false, schema.GolemStatusUpdateBody{}
	}
	// Sucess case - golem statusInfo gotten successfully
	if statusInfo.IsBlocking {
		// Fail case - Cannot change status, is in blocking status
		responses.SendRes(w, responses.Golem_In_Blocking_Status, nil, currentStatus)
		return false, schema.GolemStatusUpdateBody{}
	}
	
	// Get info on status change from request body
	gotReqBody, reqBody := getRequestBodyForGolemStatusUpdate(w, r)
	if !gotReqBody {
		// Fail case -  handled by function, simply return
		return false, schema.GolemStatusUpdateBody{}
	}

	// Check that new status in list of AllowedStatuses for archetype
	isAllowed, archetypeErr := schema.IsStatusAllowedForArchetype(archetype, reqBody.NewStatus)
	if archetypeErr != nil {
		// Fail case -  error while checking for allowed
		responses.SendRes(w, responses.Generic_Failure, nil, "Internal server error occurred while checking if new status was allowed for specified golem's archetype")
		log.Error.Printf("in ChangeGolemTask encountered error: %v", archetypeErr)
		return false, schema.GolemStatusUpdateBody{}
	}
	if !isAllowed {
		// Fail state, new status not allowed
		resMsg := fmt.Sprintf("received status: %s allowed statuses for archetype %s: %v", reqBody.NewStatus, archetype, schema.GolemArchetypes[archetype].AllowedStatuses)
		responses.SendRes(w, responses.New_Status_Not_Allowed, nil, resMsg)
		return false, schema.GolemStatusUpdateBody{}
	}
	return true, reqBody
}

// HANDLER FUNCTIONS

// Handler function for the secure route: /api/v0/my/account
func AccountInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- accountInfo --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	getUserJsonString, getUserJsonStringErr := responses.JSON(userData)
	if getUserJsonStringErr != nil {
		log.Error.Printf("Error in AccountInfo, could not format thisUser as JSON. userData: %v, error: %v", userData, getUserJsonStringErr)
	}
	log.Debug.Printf("Sending response for AccountInfo:\n%v", getUserJsonString)
	responses.SendRes(w, responses.Generic_Success, userData, "")
	log.Debug.Println(log.Cyan("-- End accountInfo --"))
}

// Handler function for the secure route: GET /api/v0/my/inventories
func InventoryInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- InventoryInfo --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	getUserJsonString, getUserJsonStringErr := responses.JSON(userData)
	if getUserJsonStringErr != nil {
		log.Error.Printf("Error in InventoryInfo, could not format thisUser as JSON. userData: %v, error: %v", userData, getUserJsonStringErr)
	}
	log.Debug.Printf("Sending response for InventoryInfo:\n%v", getUserJsonString)
	responses.SendRes(w, responses.Generic_Success, userData.Inventories, "")
	log.Debug.Println(log.Cyan("-- End InventoryInfo --"))
}

// Handler function for the secure route: GET /api/v0/my/golems
func GetGolems(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- GetGolems --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	responses.SendRes(w, responses.Generic_Success, userData.Golems, "")
	log.Debug.Println(log.Cyan("-- End GetGolems --"))
}

// Handler function for the secure route: GET /api/v0/my/golems/{archetype}
func GetGolemsByArchetype(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- GetGolemsByArchetype --"))
	route_vars := mux.Vars(r)
	archetype := trimTrailingS(route_vars["archetype"])
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	filteredList := schema.FilterGolemListByArchetype(userData.Golems, archetype)
	// getInvokerJsonString, getInvokerJsonStringErr := responses.JSON(filteredList)
	// if getInvokerJsonStringErr != nil {
	// 	log.Error.Printf("Error in GetInvGetGolemsByArchetypeokers, could not format invokers as JSON. invokers: %v, error: %v", userData, getInvokerJsonStringErr)
	// }
	// log.Debug.Printf("Sending response for GetGolemsByArchetype:\n%v", getInvokerJsonString)
	responses.SendRes(w, responses.Generic_Success, filteredList, "")
	log.Debug.Println(log.Cyan("-- End GetGolemsByArchetype --"))
}

// Handler function for the secure route: GET /api/v0/my/golems/info/{symbol}
func GolemInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- InvokerInfo --"))
	route_vars := mux.Vars(r)
	symbol := route_vars["symbol"]
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	// Find golem with symbol
	for i := range userData.Golems {
		if strings.EqualFold(userData.Golems[i].Symbol, symbol) {
			// Found
			responses.SendRes(w, responses.Generic_Success, userData.Golems[i], "")
			return
		}
	}
	// Not found
	responses.SendRes(w, responses.No_Golem_Found, nil, "")
	log.Debug.Println(log.Cyan("-- End InvokerInfo --"))
}

// Handler function for the secure route: GET /api/v0/my/rituals
func ListRituals(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- ListRituals --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	var responseData []schema.Ritual
	for _, ritual := range userData.KnownRituals {
		responseData = append(responseData, schema.Rituals[ritual])
	}
	responses.SendRes(w, responses.Generic_Success, responseData, "")
	log.Debug.Println(log.Cyan("-- End ListRituals --"))
}

// Handler function for the secure route: GET /api/v0/my/rituals/{ritual}
func GetRitualInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- GetRitualInfo --"))
	// Get ritual from route
	route_vars := mux.Vars(r)
	ritual := route_vars["ritual"]
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	responseData, ok := schema.Rituals[ritual]
	if !ok {
		// Fail case - no ritual found
		responses.SendRes(w, responses.No_Such_Ritual, nil, "")
	}
	knowsRitual := doesUserKnowRitual(userData, ritual)
	if !knowsRitual {
		responses.SendRes(w, responses.Ritual_Not_Known, nil, "")
		return
	}
	// Success case
	responses.SendRes(w, responses.Generic_Success, responseData, "")
	log.Debug.Println(log.Cyan("-- End GetRitualInfo --"))
}

// Handler function for the secure route: POST /api/v0/my/rituals/summon-invoker
func NewInvoker(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- NewInvoker --"))
	OK, userData, udb, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	success := createNewGolemInDB(w, r, udb, userData, "invoker", gamevars.Starting_Location, "summon-invoker", "invoking", gamevars.Capacity_Invoker)
	if !success {
		return // Failure states handled by createNewGolemInDB, simply return
	}
	log.Debug.Println(log.Cyan("-- End NewInvoker --"))
}

// Handler function for the secure route: POST /api/v0/my/rituals/summon-harvester
func NewHarvester(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- NewHarvester --"))
	OK, userData, udb, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	success := createNewGolemInDB(w, r, udb, userData, "harvester", gamevars.Starting_Location, "summon-harvester", "idle", gamevars.Capacity_Harvester)
	if !success {
		return // Failure states handled by createNewGolemInDB, simply return
	}
	log.Debug.Println(log.Cyan("-- End NewHarvester --"))
}

// Handler function for the secure route: POST /api/v0/my/rituals/summon-courier
func NewCourier(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- NewCourier --"))
	OK, userData, udb, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	success := createNewGolemInDB(w, r, udb, userData, "courier", gamevars.Starting_Location, "summon-courier", "idle", gamevars.Capacity_Courier)
	if !success {
		return // Failure states handled by createNewGolemInDB, simply return
	}
	log.Debug.Println(log.Cyan("-- End NewCourier --"))
}

// Handler function for the secure route: PUT /api/v0/my/invokers/{symbol}
func ChangeGolemTask(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- ChangeGolemTask --"))
	route_vars := mux.Vars(r)
	symbol := route_vars["symbol"]
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	// Find golem with symbol
	found, golemIndex := schema.FindIndexOfGolemWithSymbol(userData.Golems, symbol)
	if !found {
		// Not Found
		responses.SendRes(w, responses.No_Golem_Found, nil, "")
		return
	}
	targetGolem := &userData.Golems[golemIndex]
	currentStatus := targetGolem.Status
	archetype := targetGolem.Archetype

	// Check golem for blocking status, verify new status is allowed based on archetype
	changeAllowed, reqBody := checkStatusChangeAllowedAndGetReqBody(w, r, currentStatus, archetype)
	if !changeAllowed {
		return // Fail state, handled by func, return
	}
	// Success state, new status is allowed, complete changes based on request body
	executeGolemStatusChange(w, r, reqBody, &userData, targetGolem)
	log.Debug.Println(log.Cyan("-- End ChangeGolemTask --"))
}