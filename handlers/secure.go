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
	"github.com/brct-james/guild-golems/clearinghouse"
	"github.com/brct-james/guild-golems/gamelogic"
	"github.com/brct-james/guild-golems/gamevars"
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/metrics"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/responses"
	"github.com/brct-james/guild-golems/schema"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

// HELPER FUNCTIONS

// ENUM for handler context
type HandlerResponseKey int
const (
	UserDBContext HandlerResponseKey = iota
	WorldDBContext
)

// Gets ResourceNode for target_node if in the locale specified by location_symbol
func getTargetResNodeFromLocale(w http.ResponseWriter, r *http.Request, location_symbol string, target_node string) (bool, schema.ResourceNode) {
	// Get wdb
	wdbSuccess, wdb := GetWdbFromCtx(w, r)
	if !wdbSuccess {
		log.Debug.Printf("Could not get wdb from ctx")
		return false, schema.ResourceNode{} // Fail state, could not get wdb, handled by func - simply return
	}
	// Success state, got wdb
	// Get locale data from db
	cur_locale := schema.Locales[location_symbol]
	// Find resNodes for relevant locale
	if len(cur_locale.ResourceNodeSymbols) < 1 {
		// Fail case, no resNodes found
		log.Error.Printf("Golem has no available resource nodes! location_symbol: %s | target_node: %s | Locale: %v", location_symbol, target_node, cur_locale)
		responses.SendRes(w, responses.No_Resource_Nodes_At_Location, nil, "")
		return false, schema.ResourceNode{}
	}
	foundTargetNode, res_node := getResNodeDataIfTargetNodeInCurLocaleNodes(w, r, wdb, cur_locale, target_node)
	return foundTargetNode, res_node
}

// Gets route for target_route if in the locale specified by location_symbol
func getTargetRouteFromLocale(w http.ResponseWriter, r *http.Request, location_symbol string, target_route string) (bool, schema.Route) {
	// Get wdb
	wdbSuccess, wdb := GetWdbFromCtx(w, r)
	if !wdbSuccess {
		log.Debug.Printf("Could not get wdb from ctx")
		return false, schema.Route{} // Fail state, could not get wdb, handled by func - simply return
	}
	// Success state, got wdb
	// Get locale data from db
	cur_locale := schema.Locales[location_symbol]
	// Find routes for relevant locale
	if len(cur_locale.RouteSymbols) < 1 {
		// Fail case, no routes found
		log.Error.Printf("Golem has no available routes! location_symbol: %s | target_route: %s | Locale: %v", location_symbol, target_route, cur_locale)
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
	cur_route := schema.Routes[target_route_symbol]
	return true, cur_route
}

var lockedGolems map[string]map[string]struct{} = make(map[string]map[string]struct{}) // "username": "golemSymbol"

func checkGolemLock(w http.ResponseWriter, username string, symbol string) (bool) {
	// Check for golem lock
	log.Debug.Printf("checkGolemLock, usersymb: %s %s lockedGolems: %v", username, symbol, lockedGolems)
	if _, ok := lockedGolems[username][symbol]; ok {
		// Locked
		log.Debug.Printf("checkGolemLock: LOCKED")
		responses.SendRes(w, responses.Golem_Locked_For_Editing, nil, "")
		return true
	}
	// Lock it ourselves
	if _, ok := lockedGolems[username]; !ok {
		// No entry for username, create one
		lockedGolems[username] = make(map[string]struct{})
	}
	lockedGolems[username][symbol] = struct{}{}
	return false
}

func unlockGolem(username string, symbol string) {
	// Unlock golem
	if len(lockedGolems[username]) > 1 {
		// Just delete entry for targetGolem
		log.Debug.Printf("unlockGolem, onlytargetgol %s %s len(lockedGolems[username]): %d, lockedGolems[username][symbol]: %v", username, symbol, len(lockedGolems[username]), lockedGolems[username][symbol])
		delete(lockedGolems[username], symbol)
		log.Debug.Printf("%d %v", len(lockedGolems[username]), lockedGolems[username][symbol])
	} else {
		log.Debug.Printf("unlockGolem, entireusername %s %s len(lockedGolems[username]): %d", username, symbol, len(lockedGolems[username]))
		// Delete user's entire entry as this is the last entry
		delete(lockedGolems, username)
	}
	log.Debug.Printf("unlockGolem, UNLOCKED usersym: %s %s lockedGolems: %v", username, symbol, lockedGolems)
}

func executeGolemStatusChange(w http.ResponseWriter, r *http.Request, reqBody schema.GolemStatusUpdateBody, userData schema.User, targetGolem schema.Golem) {
	switch reqBody.NewStatus {
	case "idle":
		targetGolem.Status = "idle"
		targetGolem.StatusDetail = ""
		savedToDb := GetUDBAndSaveUserToDB(w, r, userData)
		unlockGolem(userData.Username, targetGolem.Symbol)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, targetGolem), "")
	case "harvesting":
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		nodeSymbolInInstructions, target_node := stringKeyInMap("node_symbol", statusInstructions)
		if !nodeSymbolInInstructions {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			log.Debug.Printf("'node_symbol' key required for 'harvesting' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'node_symbol' key required for 'harvesting' status")
			return
		}
		// Get ResourceNodes for golem locale
		gotNode, res_node := getTargetResNodeFromLocale(w, r, targetGolem.LocationSymbol, target_node.(string))
		if !gotNode {
			unlockGolem(userData.Username, targetGolem.Symbol)
			return // Fail state, handled by func, return
		}
		// Success, found specified res node in locale and got data on it from wdb

		// Start harvesting
		targetGolem.Status = "harvesting"
		targetGolem.StatusDetail = res_node.Symbol
		// Save to DB
		savedToDb := GetUDBAndSaveUserToDB(w, r, userData)
		unlockGolem(userData.Username, targetGolem.Symbol)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, targetGolem), "")
	case "traveling":
		// Check for all expected instructions
		// Convert Instructions to map with string keys
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		routeInInstructions, target_route := stringKeyInMap("route", statusInstructions)
		if !routeInInstructions {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			log.Debug.Printf("'route' key required for 'traveling' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'route' key required for 'traveling' status")
			return
		}
		// Get routes for golem locale
		gotRoute, cur_route := getTargetRouteFromLocale(w, r, targetGolem.LocationSymbol, target_route.(string))  
		if !gotRoute {
			unlockGolem(userData.Username, targetGolem.Symbol)
			return // Fail state, handled by func, return
		}
		// Get destination from cur_route.Symbol
		log.Debug.Printf("cur_route.Symbol: %v", cur_route.Symbol)
		destinationSymbol := strings.Split(cur_route.Symbol, "|")[1]
		// Start travel
		// May set route danger as well later, to have a result calculated after travel completed
		targetGolem.Itinerary = schema.CreateOrUpdateItinerary(targetGolem.Symbol, &userData, gamelogic.CalculateArrivalTime(cur_route.TravelTime, targetGolem.Archetype).Unix(), targetGolem.LocationSymbol, destinationSymbol, cur_route.DangerLevel)
		targetGolem.Status = "traveling"
		targetGolem.StatusDetail = destinationSymbol
		// Save to DB
		unlockGolem(userData.Username, targetGolem.Symbol)
		savedToDb := GetUDBAndSaveUserToDB(w, r, userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, targetGolem), "")
	case "packing":
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		manifestInInstructions, tempManifest := stringKeyInMap("manifest", statusInstructions)
		if !manifestInInstructions {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			log.Debug.Printf("'manifest' key required for 'packing/storing' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'manifest' key required for 'packing/storing' status")
			return
		}
		if len(tempManifest.(map[string]interface{})) < 1 {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			responses.SendRes(w, responses.Blank_Manifest_Disallowed, nil, "")
			return
		}
		// Convert manifest to map[string]int from map[string]interface{}
		c, ok := tempManifest.(map[string]interface{})
		if !ok {
			// cant assert, handle error
			unlockGolem(userData.Username, targetGolem.Symbol)
			log.Error.Printf("cant assert tempManifest.(map[string]interface{}")
			responses.SendRes(w, responses.Generic_Failure, nil, "Cant assert tempManifest.(map[string]interface{}")
			return
		}
		manifest := make(map[string]int)
		for k,v := range c {
			manifest[k] = int(v.(float64))
		}

		// Get resources in locale inventory at golem location
		
		gotInv, locInv := schema.GetInventoryByKey(targetGolem.LocationSymbol, userData.Inventories)
		if !gotInv {
			// Fail case - no items in inventory at location
			unlockGolem(userData.Username, targetGolem.Symbol)
			responses.SendRes(w, responses.No_Packable_Items, nil, "")
			return
		}

		// Validate and handle packing manifest
		for symbol, quantity := range manifest {
			// Check inv for each item in manifest
			contains, amountContained := schema.DoesInventoryContain(locInv, symbol, quantity)
			if !contains {
				// Fail case - invalid manifest, specified item not contained in sufficient quantity at location
				unlockGolem(userData.Username, targetGolem.Symbol)
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
			res := schema.Resources[symbol]
			// success state
			newCapacity += res.CapacityPerUnit * float64(quantity)
		}
		if newCapacity > targetGolem.Capacity {
			// fail state, newCapacity exceeds golem max
			unlockGolem(userData.Username, targetGolem.Symbol)
			resMsg := fmt.Sprintf("newCapacity %f exceeds golem max %f", newCapacity, targetGolem.Capacity)
			responses.SendRes(w, responses.Manifest_Overflow, nil, resMsg)
			return
		}
		
		// success state - save to db

		// TODO: make packing and storing take time (so these statuses don't finish instantly)
		targetGolem.Status = "idle"
		targetGolem.StatusDetail = ""

		savedToDb := GetUDBAndSaveUserToDB(w, r, userData)
		unlockGolem(userData.Username, targetGolem.Symbol)
		if !savedToDb {
			return // Fail state, handled by func, return
		}

		responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, targetGolem), "")
	case "storing":
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		manifestInInstructions, tempManifest := stringKeyInMap("manifest", statusInstructions)
		if !manifestInInstructions {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			log.Debug.Printf("'manifest' key required for 'packing/storing' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'manifest' key required for 'packing/storing' status")
			return
		}
		if len(tempManifest.(map[string]interface{})) < 1 {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			responses.SendRes(w, responses.Blank_Manifest_Disallowed, nil, "")
			return
		}
		// Convert manifest to map[string]int from map[string]interface{}
		c, ok := tempManifest.(map[string]interface{})
		if !ok {
			// cant assert, handle error
			unlockGolem(userData.Username, targetGolem.Symbol)
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
			unlockGolem(userData.Username, targetGolem.Symbol)
			responses.SendRes(w, responses.No_Storable_Items, nil, "")
			return
		}

		// Validate and handle storing manifest
		for symbol, quantity := range manifest {
			// Check inv for each item in manifest
			contains, amountContained := schema.DoesInventoryContain(golInv, symbol, quantity)
			if !contains {
				// Fail case - invalid manifest, specified item not contained in sufficient quantity
				unlockGolem(userData.Username, targetGolem.Symbol)
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

		savedToDb := GetUDBAndSaveUserToDB(w, r, userData)
		unlockGolem(userData.Username, targetGolem.Symbol)
		if !savedToDb {
			return // Fail state, handled by func, return
		}

		responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, targetGolem), "")
	case "invoking":
		// Start invoking
		targetGolem.Status = "invoking"
		targetGolem.StatusDetail = ""
		// Save to DB
		savedToDb := GetUDBAndSaveUserToDB(w, r, userData)
		unlockGolem(userData.Username, targetGolem.Symbol)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, targetGolem), "")
	case "transacting":
		// Start transacting - body decides order type
		statusInstructions := reqBody.Instructions.(map[string]interface{})
		orderInInstructions, tempOrder := stringKeyInMap("order", statusInstructions)
		if !orderInInstructions {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			log.Debug.Printf("'order' key required for 'transacting' status")
			responses.SendRes(w, responses.Bad_Request, nil, "'order' key required for 'transacting' status")
			return
		}
		if len(tempOrder.(map[string]interface{})) < 1 {
			// Fail case
			unlockGolem(userData.Username, targetGolem.Symbol)
			log.Debug.Printf("Order cannot be blank")
			responses.SendRes(w, responses.Blank_Order_Disallowed, nil, "")
			return
		}
		// Convert order to schema.Order from map[string]interface{}
		order := schema.Order{}
		decoder, decodeErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName:"json", Result:&order})
		decoder.Decode(tempOrder)
		if decodeErr != nil {
			// Fail case, could not decode order
			unlockGolem(userData.Username, targetGolem.Symbol)
			responses.SendRes(w, responses.Could_Not_Decode_Order, nil, "")
			return
		}
		order.Type = strings.ToUpper(order.Type)
		log.Debug.Printf("Order decoded, switch order.Type: %s", order.Type)
		switch order.Type {
		case "SELL":
			// Get resources in golem inventory
			gotInv, golInv := schema.GetInventoryByKey(targetGolem.Symbol, userData.Inventories)
			if !gotInv {
				// Fail case - no items in inventory
				unlockGolem(userData.Username, targetGolem.Symbol)
				responses.SendRes(w, responses.Insufficient_Resources_Held, nil, "")
				return
			}
			log.Debug.Printf("Got inv: %v", golInv)
			// Validate and handle sell order
			contains, amountContained := schema.DoesInventoryContain(golInv, order.ItemSymbol, order.Quantity)
			if !contains {
				// Fail case - invalid order, specified item not contained in sufficient quantity
				unlockGolem(userData.Username, targetGolem.Symbol)
				resMsg := fmt.Sprintf("Item: %s, Amount contained: %d", order.ItemSymbol, amountContained)
				responses.SendRes(w, responses.Insufficient_Resources_Held, nil, resMsg)
				return
			}
			log.Debug.Printf("Validated order, amt cont: %d", amountContained)

			// Spool Market Order & Get Order Reference
			log.Debug.Printf("SPOOLING MARKET ORDER")
			targetGolem.Status = "transacting"
			targetGolem.StatusDetail = clearinghouse.Spool(order, userData.Username, targetGolem.Symbol)
			log.Debug.Printf("Reference String: %s", targetGolem.StatusDetail)

			// Save Golem to DB
			savedToDb := GetUDBAndSaveUserToDB(w, r, userData)
			if !savedToDb {
				unlockGolem(userData.Username, targetGolem.Symbol)
				return // Fail state, handled by func, return
			}
			log.Debug.Printf("Saved successfully, executing clearinghouse order")

			// Tell clearinghouse to execute order specified by reference string
			executed := clearinghouse.Execute(targetGolem.StatusDetail)
			unlockGolem(userData.Username, targetGolem.Symbol)
			if !executed {
				// Fail state, server error!
				log.Error.Printf("Clearinghouse.Execute called but could not execute, reference string: %s", targetGolem.StatusDetail)
				responses.SendRes(w, responses.Clearinghouse_Spool_Error, nil, "")
			}
			log.Debug.Printf("Executed clearinghouse order")
			// Success case
			
			// It should be safe to use UpdateGolemLinkedData here without losing the ORDER information in statusdetail - order manager works on the database directly so no race condition
			responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, targetGolem), "")
		default:
			// Invalid order type
			unlockGolem(userData.Username, targetGolem.Symbol)
			responses.SendRes(w, responses.Invalid_Order_Type, nil, "")
			return
		}
	default:
		// Error state, newStatus passed validation but not caught by switch statement
		//TODO: this
		unlockGolem(userData.Username, targetGolem.Symbol)
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

// Get UDB & UserInfo from Middleware
func getUdbAndUserInfo(w http.ResponseWriter, r *http.Request) (bool, rdb.Database, auth.ValidationPair) {
	// Get udb from context
	udb, udbErr := GetUdbFromCtx(r)
	if udbErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get UserDBContext in secureGetUser")
		responses.SendRes(w, responses.No_UDB_Context, nil, "in secureGetUser")
		return false, udb, auth.ValidationPair{}
	}
	// Get userinfoContext from validation middleware
	userInfo, userInfoErr := GetValidationFromCtx(r)
	if userInfoErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get validationpair in secureGetUser")
		userInfoErrMsg := fmt.Sprintf("userInfo is nil, check auth validation context %v:\n%v", auth.ValidationContext, r.Context().Value(auth.ValidationContext))
		responses.SendRes(w, responses.No_AuthPair_Context, nil, userInfoErrMsg)
		return false, udb, userInfo
	}
	log.Debug.Printf("Validated with username: %s and token %s", userInfo.Username, userInfo.Token)
	return true, udb, userInfo
}

// Get User from Middleware and DB
// Returns: OK, userData, udb, userAuthPair
func secureGetUser(w http.ResponseWriter, r *http.Request) (bool, schema.User, rdb.Database, auth.ValidationPair) {
	gotUdb, udb, userInfo := getUdbAndUserInfo(w, r)
	if !gotUdb {
		return false, schema.User{}, udb, userInfo // handled by func
	}
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

	// Lastly, GetGolemMapWithPublicInfo
	thisUser.Golems = schema.UpdateGolemMapLinkedData(thisUser, thisUser.Golems) 
	return true, thisUser, udb, userInfo
}

// Returns: OK, markets, wdb
func secureGetMarkets(w http.ResponseWriter, r *http.Request) (bool, map[string]schema.Market, rdb.Database) {
	nilMkt := make(map[string]schema.Market)
	// Get wdb
	wdbSuccess, wdb := GetWdbFromCtx(w, r)
	if !wdbSuccess {
		log.Debug.Printf("Could not get wdb from ctx")
		return false, nilMkt, rdb.Database{} // Fail state, could not get wdb, handled by func - simply return
	}
	// Success state, got wdb

	markets, marketsErr := gamelogic.CalculateAllMarketTicks(wdb)
	if marketsErr != nil {
		// Fail state could not calculate market ticks
		resMsg := fmt.Sprintf("calcErr: %v", marketsErr)
		responses.SendRes(w, responses.Generic_Failure, nil, resMsg)
		return false, nilMkt, wdb
	}

	return true, markets, wdb
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
	var newGolemId int = len(schema.FilterGolemMapByArchetype(userData.Golems, archetype))
	
	newGolemSymbol := fmt.Sprintf("%s-%d", schema.GolemArchetypes[archetype].Abbreviation, newGolemId)
	newGolem := schema.NewGolem(newGolemSymbol, archetype, locationSymbol, startingStatus, capacity)
	userData.Golems[newGolemSymbol] = newGolem
	saveUserErr := schema.SaveUserToDB(udb, userData)
	if saveUserErr != nil {
		// fail state - could not save
		saveUserErrMsg := fmt.Sprintf("in createNewGolemInDB | Username: %v | schema.SaveUserToDB failed, dbSaveResult: %v", userData.Username, saveUserErr)
		log.Debug.Println(saveUserErrMsg)
		responses.SendRes(w, responses.DB_Save_Failure, nil, saveUserErrMsg)
		return false
	}
	// Updated successfully
	log.Debug.Printf("Spawned new %s golem for username %s", archetype, userData.Username)
	responses.SendRes(w, responses.Generic_Success, schema.UpdateGolemLinkedData(userData, newGolem), "")
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
	saveUserErr := schema.SaveUserToDB(udb, userData)
	if saveUserErr != nil {
		// fail state - could not save
		saveUserErrMsg := fmt.Sprintf("in GetUDBAndSaveUserToDB | Username: %v | schema.SaveUserToDB failed, dbSaveResult: %v", userData.Username, saveUserErr)
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

// Handler function for the secure route: GET /api/v0/my/itineraries
func ItineraryInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- ItineraryInfo --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	getUserJsonString, getUserJsonStringErr := responses.JSON(userData)
	if getUserJsonStringErr != nil {
		log.Error.Printf("Error in ItineraryInfo, could not format thisUser as JSON. userData: %v, error: %v", userData, getUserJsonStringErr)
	}
	log.Debug.Printf("Sending response for ItineraryInfo:\n%v", getUserJsonString)
	responses.SendRes(w, responses.Generic_Success, userData.Itineraries, "")
	log.Debug.Println(log.Cyan("-- End ItineraryInfo --"))
}

// Handler function for the secure route: GET /api/v0/my/markets
func MarketInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- MarketInfo --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	ok, markets, _ := secureGetMarkets(w, r)
	if !ok {
		return // fail state, handled already
	}
	res := make(map[string]schema.Market)
	merchants := schema.FilterGolemMapByArchetype(userData.Golems, "merchant")
	for _, merchant := range merchants {
		if locale, ok := schema.Locales[merchant.LocationSymbol]; ok {
			for _, mktSymbol := range locale.MarketSymbols {
				res[mktSymbol] = markets[mktSymbol]
			}
		}
	}
	responses.SendRes(w, responses.Generic_Success, res, "")
	log.Debug.Println(log.Cyan("-- End MarketInfo --"))
}

// Handler function for the secure route: GET /api/v0/my/orders
func OrderInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- OrderInfo --"))
	gotUI, _, userInfo := getUdbAndUserInfo(w, r)
	if !gotUI {
		return // Handled by func
	}
	orders := clearinghouse.GetOrdersByUserWithStatus(userInfo.Username, "*")
	responses.SendRes(w, responses.Generic_Success, orders, "")
	log.Debug.Println(log.Cyan("-- End OrderInfo --"))
}

// Handler function for the secure route: GET /api/v0/my/orders/{status}
func GetOrdersByStatus(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- GetOrdersByStatus --"))
	route_vars := mux.Vars(r)
	status := trimTrailingS(route_vars["status"])
	gotUI, _, userInfo := getUdbAndUserInfo(w, r)
	if !gotUI {
		return // Handled by func
	}
	orders := clearinghouse.GetOrdersByUserWithStatus(userInfo.Username, status)
	responses.SendRes(w, responses.Generic_Success, orders, "")
	log.Debug.Println(log.Cyan("-- End GetOrdersByStatus --"))
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
	filteredList := schema.UpdateGolemMapLinkedData(userData, schema.FilterGolemMapByArchetype(userData.Golems, archetype))
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

// Handler function for the secure route: POST /api/v0/my/rituals/summon-merchant
func NewMerchant(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- NewMerchant --"))
	OK, userData, udb, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	success := createNewGolemInDB(w, r, udb, userData, "merchant", gamevars.Starting_Location, "summon-merchant", "idle", gamevars.Capacity_Merchant)
	if !success {
		return // Failure states handled by createNewGolemInDB, simply return
	}
	log.Debug.Println(log.Cyan("-- End NewMerchant --"))
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
	wasLocked := checkGolemLock(w, userData.Username, symbol)
	if wasLocked {
		return // handled by checkgolemlock
	}
	// Find golem with symbol
	golem, ok := userData.Golems[strings.ToUpper(symbol)]
	if !ok {
		responses.SendRes(w, responses.No_Golem_Found, nil, "")
		return
	}
	log.Debug.Printf("found targetGolem: %s", golem.Symbol)
	targetGolem := golem
	currentStatus := targetGolem.Status
	archetype := targetGolem.Archetype

	// Check golem for blocking status, verify new status is allowed based on archetype
	changeAllowed, reqBody := checkStatusChangeAllowedAndGetReqBody(w, r, currentStatus, archetype)
	if !changeAllowed {
		return // Fail state, handled by func, return
	}
	log.Debug.Printf("StatusChangeAllowed & body: %v", reqBody)
	// Success state, new status is allowed, complete changes based on request body
	executeGolemStatusChange(w, r, reqBody, userData, targetGolem)
	log.Debug.Println(log.Cyan("-- End ChangeGolemTask --"))
}