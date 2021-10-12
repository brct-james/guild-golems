// Package handlers provides handler functions for web routes
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/brct-james/guild-golems/auth"
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/responses"
	"github.com/brct-james/guild-golems/schema"
	"github.com/gorilla/mux"
)

// Helper Functions

// Attempt to save user, returns error or nil if successful
func SaveUserToDB(udb rdb.Database, userData schema.User) error {
	err := udb.SetJsonData(userData.Token, ".", userData)
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

// Attempt to get wdb from context
func GetWdbFromCtx(w http.ResponseWriter, r *http.Request) (bool, rdb.Database) {
	log.Debug.Println("Recover wdb from context")
	wdb, ok := r.Context().Value(WorldDBContext).(rdb.Database)
	if !ok {
		log.Error.Printf("Could not get WorldDBContext in LocationsOverview")
		responses.SendRes(w, responses.No_WDB_Context, nil, "in LocationsOverview")
		return false, rdb.Database{}
	}
	return true, wdb
}

// Attempt to get user from db
func publicGetUser(w http.ResponseWriter, r *http.Request, username string, token string) (bool, schema.User, rdb.Database) {
	// Get udb from context
	udb, udbErr := GetUdbFromCtx(r)
	if udbErr != nil {
		// Fail state getting context
		log.Error.Printf("Could not get UserDBContext in publicGetUser")
		responses.SendRes(w, responses.No_UDB_Context, nil, "in publicGetUser")
		return false, schema.User{}, rdb.Database{}
	}
	// Check db for user
	thisUser, userFound, getUserErr := schema.GetUserFromDB(token, udb)
	if getUserErr != nil {
		// fail state
		getErrorMsg := fmt.Sprintf("in publicGetUser, could not get from DB for username: %s, error: %v", username, getUserErr)
		responses.SendRes(w, responses.UDB_Get_Failure, nil, getErrorMsg)
		return false, schema.User{}, rdb.Database{}
	}
	if !userFound {
		// fail state - user not found
		userNotFoundMsg := fmt.Sprintf("in publicGetUser, no user found in DB with username: %s", username)
		responses.SendRes(w, responses.User_Not_Found, nil, userNotFoundMsg)
		return false, schema.User{}, rdb.Database{}
	}
	// Success case
	return true, thisUser, udb
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

// Handler function for the route: /api/v0/leaderboards
func LeaderboardDescriptions(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- LeaderboardDescriptions --"))
	boards := make([]schema.Leaderboard, 0)
	for _, b := range schema.Leaderboards {
		boards = append(boards, b)
	}
	response := schema.GetLeaderboardDescriptionResponses(boards)
	responses.SendRes(w, responses.Generic_Success, response, "")
	log.Debug.Println(log.Cyan("-- End LeaderboardDescriptions --"))
}

// Handler function for the route: /api/v0/leaderboards/{board}
func GetLeaderboards(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- GetLeaderboards --"))
	route_vars := mux.Vars(r)
	boardKey := route_vars["board"]
	board, ok := schema.Leaderboards[boardKey]
	if !ok {
		// Fail state, board not found
		responses.SendRes(w, responses.Leaderboard_Not_Found, nil, "")
		return
	}
	responses.SendRes(w, responses.Generic_Success, board, "")
	log.Debug.Println(log.Cyan("-- End GetLeaderboards --"))
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
	OK, userData, _ := publicGetUser(w, r, username, token)
	if !OK {
		return // Failure states handled by secureGetUser, simply return
	}
	// success state
	resData := schema.PublicUserInfo{
		Username: userData.Username,
		Title: userData.Title,
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
	saveUserErr := SaveUserToDB(udb, newUser)
	if saveUserErr != nil {
		// fail state - could not save
		saveUserErrMsg := fmt.Sprintf("in UsernameClaim | Username: %v | CreateNewUserInDB failed, dbSaveResult: %v", username, saveUserErr)
		log.Debug.Println(saveUserErrMsg)
		responses.SendRes(w, responses.DB_Save_Failure, nil, saveUserErrMsg)
		return
	}
	// Created successfully
	log.Debug.Printf("Generated token %s and claimed username %s", token, username)
	responses.SendRes(w, responses.Generic_Success, newUser, "")
	log.Debug.Println(log.Cyan("-- End usernameClaim --"))
}

// Handler function for the route: /api/v0/locations
func LocationsOverview(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- locationsOverview -- "))
	wdbSuccess, wdb := GetWdbFromCtx(w, r)
	if !wdbSuccess {
		return // Fail state, could not get wdb, handled by func - simply return
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