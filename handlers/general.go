// Package handlers provides handler functions for web routes
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brct-james/brct-game/log"
	"github.com/brct-james/brct-game/rdb"
	"github.com/brct-james/brct-game/responses"
	"github.com/brct-james/brct-game/schema"
)

// Handler function for the route: /
func Homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Guild Golems")
	log.Debug.Println("Hit: homepage")
}

// Handler function for the route: /api
func ApiSelection(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ApiSel")
	log.Debug.Println("Hit: apisel")
}

// Handler function for the route: /api/v0
func V0Docs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ApiDocs")
	log.Debug.Println("Hit: apidocs")
}

// Handler function for the route: /api/v0/status
func V0Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "v0Status")
	log.Debug.Println("Hit: v0Status")
}

// Handler function for the route: /api/v0/users
func UsersSummary(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "usersSummary")
	log.Debug.Println("Hit: usersSummary")
}

// Handler function for the route: /api/v0/users/{username}
func UsernameInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "usernameInfo")
	log.Debug.Println("Hit: usernameInfo")
}

// Defines the structure for the /api/v0/users/{username}/claim (POST) response
type UsernameClaimResponse struct {
	Username string `json:"username"`
	Token string `json:"token"`
}

// // Handler function for the route: /api/v0/users/{username}/claim
// func UsernameClaim(w http.ResponseWriter, r *http.Request) {
// 	if verbose {
// 		log.Verbose.Println(log.Yellow("-- usernameClaim --"))
// 		log.Verbose.Println("Recover udb from context")
// 	}
// 	if udb, ok := r.Context().Value(UserDBContext).(db.Database); ok {
// 		// Get username from route
// 		vars := mux.Vars(r)
// 		username := vars["username"]
// 		if verbose {
// 			log.Verbose.Printf("Username Requested: %s", username)
// 		}
// 		// Validate username (length & content, plus characters) & generate token if valid
// 		token, usernameValidationStatus, genTokenErr := auth.ValidateUsernameAndGenerateToken(username, udb)
// 		if usernameValidationStatus == "OK" {
// 			if genTokenErr != nil {
// 				// Error case, output to page and console if verbose
// 				genErrorMessage := fmt.Sprintf("Username: %v | GenerateTokenErr: %v", username, genTokenErr)
// 				generationErrorRes := responses.FormatResponse(5, UsernameClaimResponse{Username: username, Token: *token}, genErrorMessage)
// 				if verbose {
// 					log.Verbose.Println(genErrorMessage)
// 				}
// 				fmt.Fprint(w, generationErrorRes)
// 			} else {
// 				// Success case, attempt to create new user in db
// 				dbSaveResult := CreateNewUserInDB(udb, username, *token)
// 				if dbSaveResult == "OK" {
// 					// Success case, output to page and console if verbose
// 					if verbose {
// 						log.Verbose.Printf("Generated token %s and claimed username %s", *token, username)
// 					}
// 					res := responses.FormatResponse(1, UsernameClaimResponse{Username: username, Token: *token}, "")
// 					fmt.Fprint(w, res)
// 				} else {
// 					// Error case, output to page and console if verbose
// 					dbsaveErrorMessage := fmt.Sprintf("Username: %v | CreateNewUserInDB failed, dbSaveResult: %v", username, dbSaveResult)
// 					dbsaveErrorRes := responses.FormatResponse(4, UsernameClaimResponse{Username: username, Token: *token}, dbsaveErrorMessage)
// 					if verbose {
// 						log.Verbose.Println(dbsaveErrorMessage)
// 					}
// 					fmt.Fprint(w, dbsaveErrorRes)
// 				}
// 			}
// 		} else {
// 			// Validation fail, ouput to page and console if verbose
// 			validationErrorMessage := fmt.Sprintf("Username: %v | ValidationResponse: %v", username, usernameValidationStatus)
// 			validationErrorRes := responses.FormatResponse(3, UsernameClaimResponse{Username: username, Token: ""}, validationErrorMessage)
// 			if verbose {
// 				log.Verbose.Println(validationErrorMessage)
// 			}
// 			fmt.Fprint(w, validationErrorRes)
// 		}
// 	}
// 	if verbose {
// 		log.Verbose.Println(log.Cyan("-- End usernameClaim --"))
// 	}
// }

// // Attemmpt to create user, return value of db.CreateUser (should be "OK" or error text)
// func CreateNewUserInDB(udb db.Database, username string, token string) string {
// 	creationSuccess := db.CreateUser(udb, username, token, 0)
// 	return creationSuccess
// }

// Handler function for the route: /api/v0/locations
func LocationsOverview(w http.ResponseWriter, r *http.Request) {
	log.Debug.Println(log.Yellow("-- locationsOverview -- "))
	log.Debug.Println("Recover wdb from context")
	// Get wdb context
	if wdb, ok := r.Context().Value(WorldDBContext).(rdb.Database); ok {
		// Output world info to page
		bytes, err := wdb.GetJsonData("world", ".")
		if err != nil {
			log.Error.Printf("Could not get world from DB! Err: %v", err)
			// TODO: This should output failure state once migrated to Responses
		}
		worldData := schema.World{}
		err = json.Unmarshal(bytes, &worldData)
		if err != nil {
			log.Error.Fatalf("Could not unmarshal world json from DB: %v", err)
		}
		fmt.Fprint(w, responses.JSON(worldData))
	} else {
		log.Error.Printf("Could not get WorldDBContext in LocationsOverview")
		// TODO: This should output failure state once migrated to Responses
	}
	log.Debug.Println(log.Cyan("-- End locationsOverview -- "))
}