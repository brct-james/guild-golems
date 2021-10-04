package auth

import (
	"os"
	"regexp"

	"github.com/brct-james/brct-game/log"
	"github.com/golang-jwt/jwt"
)

// import (
// 	"context"
// 	"crypto/rand"
// 	"fmt"
// 	"io/ioutil"
// 	"math/big"
// 	"net/http"
// 	"os"
// 	"regexp"
// 	"strings"

// 	"github.com/brct-james/guild-golems/db"
// 	"github.com/brct-james/guild-golems/log"
// 	responses "github.com/brct-james/guild-golems/responses"
// 	"github.com/golang-jwt/jwt"
// 	"github.com/joho/godotenv"
// )

// Defines struct for passing around Token-Username pairs
type ValidationPair struct{
	Username string
	Token string
}

// // enum for ValidationContext
// type ValidationResponseKey int
// const (
// 	ValidationContext ValidationResponseKey = iota
// )

// validate that username meets spec
func ValidateUsername (username string) string {
	// Defines acceptable chars
	isAlphaNumeric := regexp.MustCompile(`^[A-Za-z0-9\-\_]+$`).MatchString
	if username == "" {
		return "CANT_BE_BLANK"
	} else if len(username) <= 0 {
		return "TOO_SHORT"
	} else if !isAlphaNumeric(username) {
		return "INVALID_CHARS"
	} else {
		return "OK"
	}
}

// Generates a new token based on username and gg_access_secret
func GenerateToken(username string) (string, error) {
	// Creating access token
	// Set claims for jwt
	atClaims := jwt.MapClaims{}
	atClaims["username"]=username
	// Use signing method HS256
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	log.Debug.Printf("Got GG_ACCESS_SECRET:\n%s", os.Getenv("GG_ACCESS_SECRET"))
	// Generate token using gg_access_secret
	token, err := at.SignedString([]byte(os.Getenv("GG_ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}

// // Extract Token from request header
// func ExtractToken(r *http.Request) (token *string, ok bool) {
// 	bearerToken := r.Header.Get("Authorization")
// 	strArr := strings.Split(bearerToken, " ")
// 	if len(strArr) == 2 {
// 		return &strArr[1], true
// 	}
// 	return nil, false
// }

// // Extract token from header then parse and ensure confirms to signing method, if so return decoded token
// func VerifyTokenFormatAndDecode(r *http.Request) (*jwt.Token, error) {
// 	if tokenString, ok := ExtractToken(r); ok {
// 		if verbose {
// 			log.Verbose.Printf("Token string: %s", *tokenString)
// 		}
// 		token, err := jwt.Parse(*tokenString, func(token *jwt.Token) (interface{}, error) {
// 			//Make sure the token method conforms to SigningMethodHMAC
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			// return gg_access_secret to parser for decoding
// 			return []byte(os.Getenv("GG_ACCESS_SECRET")), nil
// 		})
// 		// Pass parse errors through to calling funcs
// 		if err != nil {
// 			return nil, err
// 		}
// 		// Return decoded token
// 		return token, nil
// 	} else {
// 		// Report failure to extract token
// 		return nil, fmt.Errorf("token extraction from header failed")
// 	}
// }

// // Verify token format and decode, then extract metadata (e.g. username) and return
// func ExtractTokenMetadata(r *http.Request) (*ValidationPair, error) {
// 	// Verify format and decode
// 	token, err := VerifyTokenFormatAndDecode(r)
// 	if verbose {
// 		log.Verbose.Printf("ExtractTokenMetadata:\nToken:\n%v\nError:\n%v\n", responses.JSON(token), err)
// 	}
//   if err != nil {
//      return nil, err
//   }
// 	// ensure token.Claims is jwt.MapClaims
//   claims, ok := token.Claims.(jwt.MapClaims)
// 	if verbose {
// 		log.Verbose.Printf("claims %v ok %v\n", claims, ok)
// 		log.Verbose.Printf("token.Valid %v\n", token.Valid)
// 	}
// 	// If token valid
//   if ok && token.Valid {
// 		username := fmt.Sprintf("%s", claims["username"])
// 		if verbose {
// 			log.Verbose.Printf("username %v\n", username)
// 		}
// 		// Return token and extracted username
// 		return &ValidationPair{
// 			Token: token.Raw,
// 			Username: username,
// 		}, nil
//   }
// 	// Fail state, token invalid and/or error
//   return nil, fmt.Errorf("token invalid or token.Claims != jwt.MapClaims")
// }

// // Verify that claimed authentication details are stored in database, if so return stored username, token, and ok=true
// func AuthenticateWithDatabase(authD *ValidationPair, userDB db.Database) (username *string, token *string, ok bool) {
// 	// Get user with claimed token
// 	if dbuser, ok := db.GetUser(userDB, authD.Token); ok {
// 		if verbose {
// 			log.Verbose.Printf("AuthenticateWithDatabase, Username: %v, Token: %v\n", dbuser.Username, dbuser.Token)
// 		}
// 		return &dbuser.Username, &dbuser.Token, true
// 	} else {
// 		return nil, nil, false
// 	}
// }

// // Extract token metadata and check claimed token against database
// func ValidateUserToken(r *http.Request, userDB db.Database) (username *string, token *string, ok bool) {
// 	// Extract metadata & validate
// 	tokenAuth, err := ExtractTokenMetadata(r)
// 	if verbose {
// 		log.Verbose.Printf("ValidateUserToken:\nTokenAuth:\n%v\nError:\n%v\n", responses.JSON(tokenAuth), err)
// 	}
// 	if err != nil {
// 		return nil, nil, false
// 	}
// 	// Check against database for existing user
// 	if dbusername, dbtoken, ok := AuthenticateWithDatabase(tokenAuth, userDB); ok {
// 		// Success state, found user and matches
// 		return dbusername, dbtoken, true
// 	} else {
// 		// Fail state, did not find user
// 		return nil, nil, false
// 	}
// }

// // Generates a middleware function for handling token validation on secure routes
// func GenerateTokenValidationMiddlewareFunc(userDB db.Database) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			if verbose {
// 				log.Verbose.Println(log.Yellow("-- GenerateTokenValidationMiddlewareFunc --"))
// 			}
// 			// Validate bearer token
// 			username, token, ok := ValidateUserToken(r, userDB)
// 			if ok {
// 				// Create validation pair
// 				validationPair := ValidationPair{
// 					Username: *username,
// 					Token: *token,
// 				}
// 				if verbose {
// 					log.Verbose.Printf("validationPair:\n%v", responses.JSON(validationPair))
// 				}
// 				// Utilize context package to pass validation pair to secure routes from the middleware
// 				ctx := r.Context()
// 				ctx = context.WithValue(ctx, ValidationContext, validationPair)
// 				r = r.WithContext(ctx)
// 				// Continue serving route
// 				next.ServeHTTP(w,r)
// 			} else {
// 				// Failed to validate, return failure message
// 				w.WriteHeader(http.StatusUnauthorized)
// 				fmt.Fprint(w, responses.FormatResponse(responses.Auth_Failure, new(interface{}), ""))
// 			}
// 			if verbose {
// 				log.Verbose.Println(log.Cyan("-- End GenerateTokenValidationMiddlewareFunc --"))
// 			}
// 		})
// 	}
// }