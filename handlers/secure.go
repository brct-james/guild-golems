// Package handlers provides handler functions for web routes
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/brct-james/guild-golems/auth"
	"github.com/brct-james/guild-golems/gamelogic"
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/metrics"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/responses"
	"github.com/brct-james/guild-golems/schema"
	"github.com/brct-james/guild-golems/timecalc"
	"github.com/gorilla/mux"
)

// HELPER FUNCTIONS

// ENUM for handler context
type HandlerResponseKey int
const (
	UserDBContext HandlerResponseKey = iota
	WorldDBContext
)

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
		responses.SendRes(w, responses.No_Available_Routes, nil, "")
		log.Error.Printf("Golem has no available routes! locale_path: %s | target_route: %s | Locale: %v", locale_path, target_route, cur_locale)
		return false, schema.Route{}
	}
	foundTargetRoute, route := getRouteDataIfTargetRouteInCurLocaleRoutes(w, r, wdb, cur_locale, target_route)
	return foundTargetRoute, route
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
	route_path := fmt.Sprintf(".%s", target_route_symbol)
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
		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	case "harvesting":
		// reqBody.Instructions
		//TODO: this
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
		locale_path := fmt.Sprintf(".%s", targetGolem.LocationSymbol)
		gotRoute, cur_route := getTargetRouteFromLocale(w, r, locale_path, target_route.(string))
		if !gotRoute {
			return // Fail state, handled by func, return
		}
		// Get destination from cur_route.Symbol
		log.Debug.Printf("cur_route.Symbol: %v", cur_route.Symbol)
		destinationSymbol := strings.Split(cur_route.Symbol, "|")[1]
		// Start travel
		// May set route danger as well later, to have a result calculated after travel completed
		targetGolem.TravelInfo.ArrivalTime = timecalc.AddSecondsToTimestamp(time.Now(), cur_route.TravelTime).Unix()
		targetGolem.Status = "traveling"
		targetGolem.LocationSymbol = destinationSymbol
		// Save to DB
		savedToDb := GetUDBAndSaveUserToDB(w, r, *userData)
		if !savedToDb {
			return // Fail state, handled by func, return
		}
		responses.SendRes(w, responses.Generic_Success, targetGolem, "")
	case "invoking":
		//TODO: this
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
	// Success case
	thisUser = gamelogic.CalculateUserUpdates(thisUser)
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
func createNewGolemInDB(w http.ResponseWriter, r *http.Request, udb rdb.Database, userData schema.User, archetype string, ritualName string, startingStatus string, capacity float64) (bool) {
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
	newGolem := schema.NewGolem(newGolemSymbol, archetype, startingStatus, capacity)
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
		responses.SendRes(w, responses.New_Status_Not_Allowed, nil, "")
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
	success := createNewGolemInDB(w, r, udb, userData, "invoker", "summon-invoker", "invoking", gamelogic.Capacity_Invoker)
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
	success := createNewGolemInDB(w, r, udb, userData, "harvester", "summon-harvester", "idle", gamelogic.Capacity_Harvester)
	if !success {
		return // Failure states handled by createNewGolemInDB, simply return
	}
	log.Debug.Println(log.Cyan("-- End NewHarvester --"))
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