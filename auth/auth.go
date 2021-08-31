package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"context"
	"log"

	"github.com/brct-james/guild-golems/db"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

var verbose = false

func ValidateUsername (username string, userDB db.Database) string {
	if username == "" {
		return "CANT_BE_BLANK"
	} else if len(username) <= 0 {
		return "TOO_SHORT"
	} else if (db.GetUser(userDB, username) != db.User{}) {
		return "ALREADY_EXISTS"
	} else {
		return "OK"
	}
}

// Generates a new token based on username and access_secret
func GenerateToken(username string) (string, error) {
	var err error
	//Creating access token
	godotenvErr := godotenv.Load("secrets.env")
	if err != nil {
		log.Fatalf("Error loading secrets.env file. %v", godotenvErr)
	} else {
		fmt.Println("Loaded secrets.env file successfully")
		fmt.Println(os.Getenv("GG_ACCESS_SECRET"))
	}
	//Set claims for jwt
	atClaims := jwt.MapClaims{}
	atClaims["authorized"]=true
	atClaims["username"]=username
	//Disabled these for now. Not sure if there is a particular use case for time data in this project.
	//Including time data does mean that tokens generated at different times for the same username are unique
	//Probably not ideal as I'd rather just use token as UUID tied to username
	// atClaims["iat"] = time.Now()
	// atClaims["nbf"] = time.Now().Add(time.Second).Unix() //1s delay before key is usable
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("GG_ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}

func ExtractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure the token method conforms to SigningMethodHMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

type AccessDetails struct {
	Username string
}

func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	fmt.Printf("token %v err %v\n", token, err)
  if err != nil {
     return nil, err
  }
  claims, ok := token.Claims.(jwt.MapClaims)
	fmt.Printf("claims %v ok %v\n", claims, ok)
	fmt.Printf("token.Valid %v\n", token.Valid)
  if ok && token.Valid {
		username := fmt.Sprintf("%s", claims["username"])
		fmt.Printf("username %v\n", username)
     return &AccessDetails{
        Username:   username,
     }, nil
  }
  return nil, err
}

func FetchAuth(authD *AccessDetails, userDB db.Database) string {
	username := db.GetUserData(userDB, authD.Username, ".username")
	fmt.Printf("FetchAuth, Username: %v\n", username)
	return username.(string)
}

func ValidateUserToken(w http.ResponseWriter, r *http.Request, userDB db.Database) string {
	tokenAuth, err := ExtractTokenMetadata(r)
	fmt.Printf("tokenAuth %v err %v\n", tokenAuth, err)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "{\"error\":{\"message\": \"Token was invalid or missing from request. Did you confirm sending the token as an authorization header?\",\"code\":40101}}")
		// json.NewEncoder(w).Encode("Missing auth token")
		return ""
	}
	return FetchAuth(tokenAuth, userDB)
}

// Generates a middleware function for handling token validation on certain routes
func GenerateTokenValidationMiddlewareFunc(userDB db.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Hit token validation")
			validatedUsername := ValidateUserToken(w, r, userDB)
			fmt.Printf("validatedUsername? %v\n", validatedUsername)
			if len(validatedUsername) > 0 {
				// Utilize context package to pass data to secure routes from the middleware
				ctx := r.Context()
				ctx = context.WithValue(ctx, "validatedUsername", validatedUsername)
				r = r.WithContext(ctx)
				next.ServeHTTP(w,r)
			}
		})
	}
}