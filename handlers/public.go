// Package handlers provides handler functions for web routes
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/brct-james/brct-game/auth"
	"github.com/brct-james/brct-game/log"
	"github.com/brct-james/brct-game/rdb"
	"github.com/brct-james/brct-game/responses"
	"github.com/brct-james/brct-game/schema"
	"github.com/gorilla/mux"
)

// Helper Functions

// Attemmpt to save user, returns error or nil if successful
func SaveUserToDB(udb rdb.Database, token string, userData schema.User) error {
	err := udb.SetJsonData(token, ".", userData)
	// creationSuccess := rdb.CreateUser(udb, username, token, 0)
	return err
}

// Attempt to get udb from context, return udb, nil if successful else return rdb.Database{}, nil
func GetUdbFromCtx(r *http.Request) (rdb.Database, error) {
	log.Debug.Println("Recover udb from context")
	udb, ok := r.Context().Value(UserDBContext).(rdb.Database)
	if !ok {
		return rdb.Database{}, errors.New("could not get UserDBContext")
	}
	return udb, nil
}

// Handler Functions

// Handler function for the route: /
func Homepage(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- Homepage --"))
	responses.SendRes(w, responses.Unimplemented, nil, "Homepage")
	log.Debug.Println(log.Cyan("-- End Homepage --"))
}

// Handler function for the route: /api
func ApiSelection(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- apiSel --"))
	responses.SendRes(w, responses.Unimplemented, nil, "apiSel")
	log.Debug.Println(log.Cyan("-- End apiSel --"))
}

// Handler function for the route: /api/v0
func V0Status(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- v0Status --"))
	responses.SendRes(w, responses.Unimplemented, nil, "v0Status")
	log.Debug.Println(log.Cyan("-- End v0Status --"))
}

// Handler function for the route: /api/v0/users
func UsersSummary(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- usersSummary --"))
	responses.SendRes(w, responses.Unimplemented, nil, "usersSummary")
	log.Debug.Println(log.Cyan("-- End usersSummary --"))
}

// Handler function for the route: /api/v0/users/{username}
func UsernameInfo(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- usernameInfo --"))
	udb, udbErr := GetUdbFromCtx(r)
	if udbErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get UserDBContext in UsernameInfo")
		responses.SendRes(w, responses.No_UDB_Context, nil, "in UsernameInfo")
		return
	}
	// Get username from route
	route_vars := mux.Vars(r)
	username := route_vars["username"]
	log.Debug.Printf("UsernameInfo Requested for: %s", username)
	// Get username info from DB
	token, genTokenErr := auth.GenerateToken(username)
	if genTokenErr != nil {
		// fail state
		log.Important.Printf("in UsernameInfo: Attempted to generate token using username %s but was unsuccessful with error: %v", username, genTokenErr)
		genErrorMsg := fmt.Sprintf("Could not get, failed to convert username to DB token. Username: %v | GenerateTokenErr: %v", username, genTokenErr)
		responses.SendRes(w, responses.Generate_Token_Failure, nil, genErrorMsg)
		return
	}
	userData, userFound, getError := schema.GetUserFromDB(token, udb)
	if getError != nil {
		// fail state
		getErrorMsg := fmt.Sprintf("in UsernameInfo, could not get from DB for username: %s, error: %v", username, getError)
		responses.SendRes(w, responses.UDB_Get_Failure, nil, getErrorMsg)
		return
	}
	if !userFound {
		// fail state - user not found
		userNotFoundMsg := fmt.Sprintf("in UsernameInfo, no user found in DB with username: %s", username)
		responses.SendRes(w, responses.User_Not_Found, nil, userNotFoundMsg)
		return
	}
	// success state
	resData := schema.PublicUserInfo{
		Username: userData.Username,
		Tagline: userData.Tagline,
		Coins: userData.Coins,
		UserSince: userData.UserSince,
	}
	responses.SendRes(w, responses.Generic_Success, resData, "")
	log.Debug.Println(log.Cyan("-- End usernameInfo --"))
}

// Handler function for the route: /api/v0/users/{username}/claim
func UsernameClaim(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- usernameClaim --"))
	log.Debug.Println("Recover udb from context")
	udb, udbErr := GetUdbFromCtx(r)
	if udbErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get UserDBContext in UsernameClaim")
		responses.SendRes(w, responses.No_UDB_Context, nil, "in UsernameClaim")
		return
	}
	// Get username from route
	route_vars := mux.Vars(r)
	username := route_vars["username"]
	log.Debug.Printf("Username Requested: %s", username)
	// Validate username (length & content, plus characters)
	usernameValidationStatus := auth.ValidateUsername(username)
	if usernameValidationStatus != "OK" {
		// fail state
		validationErrorMessage := fmt.Sprintf("in UsernameClaim: Username: %v | ValidationResponse: %v", username, usernameValidationStatus)
		log.Debug.Println(validationErrorMessage)
		responses.SendRes(w, responses.Username_Validation_Failure, nil, validationErrorMessage)
		return
	}
	// generate token
	token, genTokenErr := auth.GenerateToken(username)
	if genTokenErr != nil {
		// fail state
		log.Important.Printf("in UsernameClaim: Attempted to generate token using username %s but was unsuccessful with error: %v", username, genTokenErr)
		genErrorMsg := fmt.Sprintf("Username: %v | GenerateTokenErr: %v", username, genTokenErr)
		responses.SendRes(w, responses.Generate_Token_Failure, nil, genErrorMsg)
		return
	}
	// check DB for existing user
	userExists, dbGetError := schema.CheckForExistingUser(token, udb)
	if dbGetError != nil {
		// fail state - db error
		dbGetErrorMsg := fmt.Sprintf("in UsernameClaim | Username: %v | UDB Get Error: %v", username, dbGetError)
		log.Debug.Println(dbGetErrorMsg)
		responses.SendRes(w, responses.UDB_Get_Failure, nil, dbGetErrorMsg)
		return
	}
	if userExists {
		// fail state - user already exists
		validationFailMsg := fmt.Sprintf("in UsernameClaim | Username: %v | Reason: USER_ALREADY_EXISTS", username)
		log.Debug.Println(validationFailMsg)
		responses.SendRes(w, responses.Username_Validation_Failure, nil, validationFailMsg)
		return
	}
	// create new user in DB
	newUser := schema.NewUser(token, username)
	saveUserErr := SaveUserToDB(udb, token, newUser)
	if saveUserErr != nil {
		// fail state - could not save
		saveUserErrMsg := fmt.Sprintf("in UsernameClaim | Username: %v | CreateNewUserInDB failed, dbSaveResult: %v", username, saveUserErr)
		log.Debug.Println(saveUserErrMsg)
		responses.SendRes(w, responses.DB_Save_Failure, nil, saveUserErrMsg)
		return
	}
	// Created successfully
	log.Debug.Printf("Generated token %s and claimed username %s", token, username)
	responses.SendRes(w, 1, newUser, "")
	log.Debug.Println(log.Cyan("-- End usernameClaim --"))
}

// Handler function for the route: /api/v0/locations
func LocationsOverview(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- locationsOverview -- "))
	log.Debug.Println("Recover wdb from context")
	// Get wdb context
	wdb, ok := r.Context().Value(WorldDBContext).(rdb.Database)
	if !ok {
		log.Error.Printf("Could not get WorldDBContext in LocationsOverview")
		responses.SendRes(w, responses.No_WDB_Context, nil, "in LocationsOverview")
		return
	}
	// Output world info to page
	bytes, err := wdb.GetJsonData("world", ".")
	if err != nil {
		log.Error.Printf("Could not get world from DB! Err: %v", err)
		responses.SendRes(w, responses.WDB_Get_Failure, nil, "in LocationsOverview")
		return
	}
	worldData := schema.World{}
	err = json.Unmarshal(bytes, &worldData)
	if err != nil {
		log.Error.Printf("Could not unmarshal world json from DB: %v", err)
		responses.SendRes(w, responses.JSON_Unmarshal_Error, nil, "in LocationsOverview")
		return
	}
	responses.SendRes(w, responses.Generic_Success, worldData, "")
	log.Debug.Println(log.Cyan("-- End locationsOverview -- "))
}