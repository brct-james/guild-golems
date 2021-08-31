package db

import (
	"encoding/json"
	"log"
)

type User struct {
	Username string `json:"username" binding:"required"`
	Token string `json:"token" binding:"required"`
	Coins uint `json:"coins" binding:"required"`
}

// Set user
func SetUser (db Database, username string, token string, coins uint) {
	user := User{
		Username: username,
		Token: token,
		Coins: coins,
	}
	JsonSetData(db.Rejson, username, user)
}

// Get user and unmarshall into User
func GetUser(db Database, username string) User {
	method := "GetUser"
	dataJSON := JsonGetData(db.Rejson, username)
	readData := User{}
	err := json.Unmarshal(dataJSON, &readData)
	if err != nil {
		log.Printf("%s: Failed to JSON Unmarshal. dataJSON: %s", method, dataJSON)
		return User{}
	}

	if verbose {
		log.Printf("%s: Unmarshalled from redis : %#v\n", method, readData)
	}
	return readData
}