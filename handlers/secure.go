// Package handlers provides handler functions for web routes
package handlers

import (
	"context"
	"net/http"

	"github.com/brct-james/brct-game/log"
	"github.com/brct-james/brct-game/rdb"
)

// ENUM for handler context
type HandlerResponseKey int
const (
	UserDBContext HandlerResponseKey = iota
	WorldDBContext
)

// Access middleware context using, for example:
// if udb, ok := r.Context().Value(UserDBContext).(rdb.Database); ok {}

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

// // Handler function for the secure route: /api/v0/my/account
// func AccountInfo(w http.ResponseWriter, r *http.Request) {
// 	log.Debug.Println(log.Yellow("-- accountInfo --"))
// 	log.Debug.Println("Recover udb from context")
// 	// Get udb from context
// 	if udb, ok := r.Context().Value(UserDBContext).(rdb.Database); ok {
// 		// Get userinfoContext from validation middleware
// 		if userInfo, ok := r.Context().Value(auth.ValidationContext).(auth.ValidationPair); ok {
// 			log.Debug.Printf("Validated with username: %s and token %s", userInfo.Username, userInfo.Token)
// 			// Check db for user
// 			if thisUser, ok := rdb.GetUser(udb, userInfo.Token); ok {
// 				// Success case, Output user info to page
// 				log.Debug.Printf("Got response from db for GetUser:\n%v", responses.JSON(thisUser))
// 				fmt.Fprint(w, responses.JSON(thisUser))
// 			} else {
// 				// Failure case, Output error to page
// 				log.Debug.Printf("failed to GetUser from db with username: %s", userInfo.Username)
// 				fmt.Fprint(w, responses.JSON(rdb.User{}))
// 			}
// 		} else {
// 			log.Important.Printf("userInfo is nil, check auth validation context %v:\n%v", auth.ValidationContext, r.Context().Value(auth.ValidationContext))
// 			fmt.Fprint(w, responses.FormatResponse(0, new(interface{}), ""))
// 		}
// 	}
// 	log.Debug.Println(log.Cyan("-- End accountInfo --"))
// }