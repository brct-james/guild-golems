package tokengen

import (
	"os"

	"github.com/brct-james/guild-golems/log"
	"github.com/golang-jwt/jwt"
)

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
