// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
)

// Defines a user which has Name, Symbol, Description
type User struct {
	Token string `json:"token" binding:"required"`
	Username string `json:"username" binding:"required"`
	Tagline string `json:"tagline" binding:"required"`
	Coins uint64 `json:"coins" binding:"required"`
	UserSince int64 `json:"user-since" binding:"required"`
}

// Defines the public User info for the /users/{username} endpoint
type PublicUserInfo struct {
	Username string `json:"username" binding:"required"`
	Tagline string `json:"tagline" binding:"required"`
	Coins uint64 `json:"coins" binding:"required"`
	UserSince int64 `json:"user-since" binding:"required"`
}

func NewUser(token string, username string) User {
	return User{
		Token: token,
		Username: username,
		Tagline: "",
		Coins: 0,
		UserSince: time.Now().Unix(),
	}
}

// Check DB for existing user with given token and return bool for if exists, and error if error encountered
func CheckForExistingUser (token string, udb rdb.Database) (bool, error) {
	// Get user
	_, getError := udb.GetJsonData(token, ".")
	if getError != nil {
		if fmt.Sprint(getError) != "redis: nil" {
			// error
			return false, getError
		}
		// user not found
		return false, nil
	}
	// Got successfully
	return true, nil
}

// Get user from DB, bool is user found
func GetUserFromDB (token string, udb rdb.Database) (User, bool, error) {
	// Get user json
	uJson, getError := udb.GetJsonData(token, ".")
	if getError != nil {
		if fmt.Sprint(getError) != "redis: nil" {
			// user not found
			return User{}, false, nil
		}
		// error
		return User{}, false, getError
	}
	// Got successfully, unmarshal
	uData := User{}
	unmarshalErr := json.Unmarshal(uJson, &uData)
	if unmarshalErr != nil {
		log.Error.Fatalf("Could not unmarshal user json from DB: %v", unmarshalErr)
		return User{}, false, unmarshalErr
	}
	return uData, true, nil
}