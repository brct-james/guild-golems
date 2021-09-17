package auth

import (
	// 	"context"
	// 	"fmt"
	// 	"net/http"
	"os"
	// 	"regexp"
	"crypto/rand"
	"math/big"

	// 	"github.com/brct-james/brct-game/db"
	"github.com/brct-james/brct-game/filemngr"
	"github.com/brct-james/brct-game/log"

	// 	responses "github.com/brct-james/brct-game/responses"
	// 	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

// Creates or updates the GG_ACCESS_SECRET value in secrets.env
func CreateOrUpdateAuthSecretInFile() {
	// Ensure exists
	filemngr.Touch("secrets.env")
	// Read file to lines array splitting by newline
	lines, readErr := filemngr.ReadFileToLineSlice("secrets.env")
	if readErr != nil {
		// Auth is mission-critical, using Fatal
		log.Error.Fatalf("Could not read lines from secrets.env. Err: %v", readErr)
	}

	// Securely generate new 64 character secret
	newSecret, generationErr := GenerateRandomSecureString(64)
	if generationErr != nil {
		log.Error.Fatalf("Could not generate secure string: %v", generationErr)
	}
	secretString :=  "GG_ACCESS_SECRET=" + string(newSecret)
	log.Debug.Printf("New Secret Generated: %s", secretString)
	
	// Search existing file for secret identifier
	found, i := filemngr.KeyInSliceOfLines("GG_ACCESS_SECRET=", lines)
	if found {
		// Update existing secret
		lines [i] = secretString
	} else {
		// Create secret in env file since could not find one to update
		// If empty file then replace 1st line else append to end
		log.Debug.Printf("Creating new secret in env file. secrets.env[0] == ''? %v", lines[0] == "")
		if lines[0] == "" {
			log.Debug.Printf("Blank secrets.env, replacing line 0")
			lines[0] = secretString
		} else {
			log.Debug.Printf("Not blank secrets.env, appending to end")
			lines = append(lines, secretString)
		}
	}
	
	// Join and write out
	writeErr := filemngr.WriteLinesToFile("secrets.env", lines)
	if writeErr != nil {
		log.Error.Fatalf("Could not write secrets.net: %v", writeErr)
	}
	log.Info.Println("Wrote auth secret to secrets.env")
}

// Generate random string of n characters
func GenerateRandomSecureString(n int) (string, error) {
	const allowed = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowed))))
		if err != nil {
			return "", err
		}
		ret[i] = allowed[num.Int64()]
	}
	return string(ret), nil
}


// Load secrets.env file to environment
func LoadSecretsToEnv() {
	godotenvErr := godotenv.Load("secrets.env")
	if godotenvErr != nil {
		// Loading secrets is mission-critical, fatal
		log.Error.Fatalf("Error loading secrets.env file. %v", godotenvErr)
	} else {
		log.Info.Println("Loaded secrets.env file successfully")
		log.Debug.Printf("GG_ACCESS_SECRET: %s", os.Getenv("GG_ACCESS_SECRET"))
	}
}

// // Used by UsernameClaim handler to validate username meets spec before attempting user creation
// func ValidateUsernameAndGenerateToken (username string, userDB db.Database) (token *string, usernameValidationStatus string, genTokenErr error) {
// 	// Defines acceptable chars
// 	isAlphaNumeric := regexp.MustCompile(`^[A-Za-z0-9\-\_]+$`).MatchString
// 	if username == "" {
// 		return nil, "CANT_BE_BLANK", nil
// 	} else if len(username) <= 0 {
// 		return nil, "TOO_SHORT", nil
// 	} else if !isAlphaNumeric(username) {
// 		return nil, "INVALID_CHARS", nil
// 	} else {
// 		// Generate a token using username and check if already exists in db
// 		token, err := GenerateToken(username)
// 		if err != nil {
// 			// GenerateToken had error, pass up to calling func
// 			log.Warning.Printf("ValidateUsername: Attempted to generate token using username %s but as unsuccessful with error: %v", username, err)
// 			return nil, "OK", err
// 		}
// 		// Get user and see if already exists
// 		if _, ok := db.GetUser(userDB, *token); ok {
// 			return nil, "ALREADY_EXISTS", nil
// 		} else {
// 			// Could not get, doesn't already exist, pass OK and return token
// 			return token, "OK", nil
// 		}
// 	}
// }

// // Generates a new token based on username and gg_access_secret
// func GenerateToken(username string) (*string, error) {
// 	// Creating access token
// 	// Set claims for jwt
// 	atClaims := jwt.MapClaims{}
// 	atClaims["username"]=username
// 	// Use signing method HS256
// 	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
// 	if verbose {
// 		log.Verbose.Printf("Got GG_ACCESS_SECRET:\n%s", os.Getenv("GG_ACCESS_SECRET"))
// 	}
// 	// Generate token using gg_access_secret
// 	token, err := at.SignedString([]byte(os.Getenv("GG_ACCESS_SECRET")))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &token, nil
// }

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

// // Defines struct for passing around Token-Username pairs
// type ValidationPair struct{
// 	Username string
// 	Token string
// }

// // enum for ValidationContext
// type ValidationResponseKey int
// const (
// 	ValidationContext ValidationResponseKey = iota
// )

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