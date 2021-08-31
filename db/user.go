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

// Create user
func CreateUser (db Database, username string, token string, coins uint) string {
	user := User{
		Username: username,
		Token: token,
		Coins: coins,
	}
	return JsonSetData(db.Rejson, username, ".", user)
}

// Update user
func UpdateUser (db Database, username string, token string, path string, newValue interface{}) {
	userToken := GetUserData(db, username, ".token")
	if (userToken != nil) {
		if (userToken == token) {
			dbSaveResult := JsonSetData(db.Rejson, username, path, newValue)
			if dbSaveResult == "OK" {
				if verbose {
					log.Printf("Updated user data %v at path %v", username, path)
				}
			} else {
				if verbose {
					log.Printf("Could not update user (%v) at path %v. Error saving to DB. dbSaveResult : %v", username, path, dbSaveResult)
				}
			}
		} else {
			if  verbose {
				log.Printf("UpdateUser: %s attempted update with incorrect token %s", username, token)
			}
		}
	} else {
		if verbose {
			log.Printf("UpdateUser: %s attempted update but no user by the name exists", username)
		}
	}
}

// Get user and unmarshall into User
func GetUser(db Database, username string) User {
	method := "GetUser"
	dataJSON := JsonGetData(db.Rejson, username, ".")
	readData := User{}
	err := json.Unmarshal(dataJSON, &readData)
	if err != nil {
		if verbose {
			log.Printf("%s: Failed to JSON Unmarshal. dataJSON: %s", method, dataJSON)
		}
		return User{}
	}

	if verbose {
		log.Printf("%s: Unmarshalled from redis : %#v\n", method, readData)
	}
	return readData
}

// Gett user data at path
func GetUserData(db Database, username string, path string) interface{} {
	method := "GetUserData"
	dataJSON := JsonGetData(db.Rejson, username, path)
	var readData interface{}
	err := json.Unmarshal(dataJSON, &readData)
	if err != nil {
		if verbose {
			log.Printf("%s: Failed to JSON Unmarshal. dataJSON: %s", method, dataJSON)
		}
		return nil
	}

	if verbose {
		log.Printf("%s: Unmarshalled from redis : %#v\n", method, readData)
	}
	return readData
}