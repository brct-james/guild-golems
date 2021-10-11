// Package handlers provides handler functions for web routes
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/brct-james/guild-golems/auth"
	"github.com/brct-james/guild-golems/gamelogic"
	"github.com/brct-james/guild-golems/log"
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

// Access middleware context using, for example:
// [DEPRECATED] if udb, ok := r.Context().Value(UserDBContext).(rdb.Database); ok {}
// [NEW] Wrote GetUdbFromCtx(r) for this purpose, use like:
	// udb, udbErr := GetUdbFromCtx(r)
	// if udbErr != nil {
	// 	// Fail state getting context
	// 	log.Error.Printf("Could not get UserDBContext in UsernameInfo")
	// 	responses.SendRes(w, responses.No_UDB_Context, nil, "in UsernameInfo")
	// 	return
	// }
// Similarly wrote the below for getting validation context from auth middleware

// Attempt to get validation context
func GetValidationFromCtx(r *http.Request) (auth.ValidationPair, error) {
	log.Debug.Println("Recover validationpair from context")
	userInfo, ok := r.Context().Value(auth.ValidationContext).(auth.ValidationPair)
	if !ok {
		return auth.ValidationPair{}, errors.New("could not get ValidationPair")
	}
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
	thisUser = gamelogic.CalculateManaRegen(thisUser)
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
func createNewGolemInDB(w http.ResponseWriter, r *http.Request, udb rdb.Database, userData schema.User, archetype string, ritualName string, startingStatus string) (bool) {
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
	newGolem := schema.NewGolem(newGolemSymbol, archetype, startingStatus)
	userData.Golems = append(userData.Golems, newGolem)
	saveUserErr := SaveUserToDB(udb, userData.Token, userData)
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
	success := createNewGolemInDB(w, r, udb, userData, "invoker", "summon-invoker", "invoking")
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
	success := createNewGolemInDB(w, r, udb, userData, "harvester", "summon-harvester", "idle")
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
	currentStatus := userData.Golems[golemIndex].Status
	archetype := userData.Golems[golemIndex].Archetype
	// Found golem, check that not in blocking status
	statusInfo, ok := schema.GolemStatuses[currentStatus]
	if !ok {
		// Fail case - golem status not in list of valid statuses
		resMsg1 := fmt.Sprintf("in ChangeGolemTask, golem status %s not in list of valid statuses", currentStatus)
		responses.SendRes(w, responses.Generic_Failure, nil, resMsg1)
		return
	}
	// Sucess case - golem statusInfo gotten successfully
	if statusInfo.IsBlocking {
		// Cannot change status, is in blocking status
		responses.SendRes(w, responses.Golem_In_Blocking_Status, nil, "")
		return
	}
	
	// Get info on status change from request body
	//TODO: this
	var newStatus string = "idle" //TODO: should get from body

	// Check that new status in list of AllowedStatuses for archetype
	isAllowed, archetypeErr := schema.IsStatusAllowedForArchetype(archetype, newStatus)
	if archetypeErr != nil {
		// Fail state, error while checking for allowed
		responses.SendRes(w, responses.Generic_Failure, nil, "Internal server error occurred while checking if new status was allowed for specified golem's archetype")
		log.Error.Printf("in ChangeGolemTask encountered error: %v", archetypeErr)
		return
	}
	if !isAllowed {
		// Fail state, new status not allowed
		responses.SendRes(w, responses.New_Status_Not_Allowed, nil, "")
		return
	}
	// Success state, new status is allowed, complete changes based on request body
	//TODO: this
	switch newStatus {
	case "idle":
		//TODO: this
	case "harvesting":
		//TODO: this
	case "traveling":
		//TODO: this
	case "invoking":
		//TODO: this
	default:
		// Error state, newStatus passed validation but not caught by switch statement
		//TODO: this
	}
	responses.SendRes(w, responses.Generic_Success, userData.Golems[golemIndex], "")
	log.Debug.Println(log.Cyan("-- End ChangeGolemTask --"))
}