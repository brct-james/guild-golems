// Package gamelogic provides functions for game logic
package gamelogic

import (
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/schema"
)

// Calculates all updates to the user object based on game logic, saving to db & returns the updated user
func CalculateUserUpdates(userData schema.User) (schema.User) {
	log.Debug.Println(log.Cyan("-- Begin CalculateUserUpdates --"))
	userData = CalculateManaRegen(userData)
	userData = CalculateTravelArrived(userData)

	// Save changes to DB
	
	log.Debug.Println(log.Cyan("-- End CalculateUserUpdates --"))
	return userData
}