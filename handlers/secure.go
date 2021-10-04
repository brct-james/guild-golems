// Package handlers provides handler functions for web routes
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/brct-james/brct-game/auth"
	"github.com/brct-james/brct-game/log"
	"github.com/brct-james/brct-game/rdb"
	"github.com/brct-james/brct-game/responses"
	"github.com/brct-james/brct-game/schema"
)

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

// Handler function for the secure route: /api/v0/my/account
func AccountInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- accountInfo --"))
	log.Debug.Println("Recover udb from context")
	// Get udb from context
	udb, udbErr := GetUdbFromCtx(r)
	if udbErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get UserDBContext in AccountInfo")
		responses.SendRes(w, responses.No_UDB_Context, nil, "in AccountInfo")
		return
	}
	// Get userinfoContext from validation middleware
	userInfo, userInfoErr := GetValidationFromCtx(r)
	if userInfoErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get validationpair in AccountInfo")
		userInfoErrMsg := fmt.Sprintf("userInfo is nil, check auth validation context %v:\n%v", auth.ValidationContext, r.Context().Value(auth.ValidationContext))
		responses.SendRes(w, responses.No_AuthPair_Context, nil, userInfoErrMsg)
		return
	}
	log.Debug.Printf("Validated with username: %s and token %s", userInfo.Username, userInfo.Token)
	// Check db for user
	thisUser, userFound, getUserErr := schema.GetUserFromDB(userInfo.Token, udb)
	if getUserErr != nil {
		// fail state
		getErrorMsg := fmt.Sprintf("in accountInfo, could not get from DB for username: %s, error: %v", userInfo.Username, getUserErr)
		responses.SendRes(w, responses.UDB_Get_Failure, nil, getErrorMsg)
		return
	}
	if !userFound {
		// fail state - user not found
		userNotFoundMsg := fmt.Sprintf("in accountInfo, no user found in DB with username: %s", userInfo.Username)
		responses.SendRes(w, responses.User_Not_Found, nil, userNotFoundMsg)
		return
	}
	// Success case, Output user info to page
	getUserJsonString, getUserJsonStringErr := responses.JSON(thisUser)
	if getUserJsonStringErr != nil {
		log.Error.Printf("Error in AccountInfo, could not format thisUser as JSON. thisUser: %v, error: %v", thisUser, getUserJsonStringErr)
	}
	log.Debug.Printf("Got response from db for GetUser:\n%v", getUserJsonString)
	responses.SendRes(w, responses.Generic_Success, thisUser, "")
	log.Debug.Println(log.Cyan("-- End accountInfo --"))
}