// Package handlers provides handler functions for web routes
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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

// Handler function for the secure route: GET /api/v0/my/invokers
func GetInvokers(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- GetInvokers --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	invokers := schema.FilterGolemListByPurpose(userData.Golems, "invoker")
	getInvokerJsonString, getInvokerJsonStringErr := responses.JSON(invokers)
	if getInvokerJsonStringErr != nil {
		log.Error.Printf("Error in GetInvokers, could not format invokers as JSON. invokers: %v, error: %v", userData, getInvokerJsonStringErr)
	}
	log.Debug.Printf("Sending response for GetInvokers:\n%v", getInvokerJsonString)
	responses.SendRes(w, responses.Generic_Success, invokers, "")
	log.Debug.Println(log.Cyan("-- End GetInvokers --"))
}

// Handler function for the secure route: GET /api/v0/my/invokers/{symbol}
func InvokerInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- InvokerInfo --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	//* Placeholder/
	getUserJsonString, getUserJsonStringErr := responses.JSON(userData)
	if getUserJsonStringErr != nil {
		log.Error.Printf("Error in InvokerInfo, could not format thisUser as JSON. userData: %v, error: %v", userData, getUserJsonStringErr)
	}
	log.Debug.Printf("Sending response for InvokerInfo:\n%v", getUserJsonString)
	responses.SendRes(w, responses.Generic_Success, userData, "")
	//*/
	log.Debug.Println(log.Cyan("-- End InvokerInfo --"))
}

// Handler function for the secure route: PUT /api/v0/my/invokers/{symbol}
func ChangeInvokerTask(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- ChangeInvokerTask --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	//* Placeholder/
	getUserJsonString, getUserJsonStringErr := responses.JSON(userData)
	if getUserJsonStringErr != nil {
		log.Error.Printf("Error in ChangeInvokerTask, could not format thisUser as JSON. userData: %v, error: %v", userData, getUserJsonStringErr)
	}
	log.Debug.Printf("Sending response for ChangeInvokerTask:\n%v", getUserJsonString)
	responses.SendRes(w, responses.Generic_Success, userData, "")
	//*/
	log.Debug.Println(log.Cyan("-- End ChangeInvokerTask --"))
}

// Handler function for the secure route: GET /api/v0/my/rituals
func ListRituals(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- ListRituals --"))
	OK, userData, _, _ := secureGetUser(w, r)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	responses.SendRes(w, responses.Generic_Success, userData.KnownRituals, "")
	log.Debug.Println(log.Cyan("-- End ListRituals --"))
}

// Handler function for the secure route: GET /api/v0/my/rituals/{ritual}
func GetRitualInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- GetRitualInfo --"))
	// Get ritual from route
	route_vars := mux.Vars(r)
	ritual := route_vars["ritual"]
	var responseData schema.Ritual
	switch ritual {
	case "summon-invoker":
		responseData = schema.NewRitual("Summon Invoker", "summon-invoker", "Spend mana to summon a new invoker, who can be used to help generate even more mana.", 600)
	default:
		// Fail case - no ritual found
		responses.SendRes(w, responses.No_Such_Ritual, nil, "")
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
	success, newManaValue := gamelogic.TryManaPurchase(w, userData.Mana, 600)
	if !success {
		return // Failure states handled by TryManaPurchase, simply return
	}
	userData.Mana = newManaValue
	invokers := schema.FilterGolemListByPurpose(userData.Golems, "invoker")
	newGolem := schema.NewGolem(fmt.Sprintf("INV-%v", len(invokers)), "invoker")
	userData.Golems = append(userData.Golems, newGolem)
	saveUserErr := SaveUserToDB(udb, userData.Token, userData)
	if saveUserErr != nil {
		// fail state - could not save
		saveUserErrMsg := fmt.Sprintf("in NewInvoker | Username: %v | SaveUserToDB failed, dbSaveResult: %v", userData.Username, saveUserErr)
		log.Debug.Println(saveUserErrMsg)
		responses.SendRes(w, responses.DB_Save_Failure, nil, saveUserErrMsg)
		return
	}
	// Updated successfully
	log.Debug.Printf("Spawned new invoker for username %s", userData.Username)
	responses.SendRes(w, responses.Generic_Success, newGolem, "")
	log.Debug.Println(log.Cyan("-- End NewInvoker --"))
}