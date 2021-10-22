// Package gamelogic provides functions for game logic
package gamelogic

import (
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/schema"
)

// Calculates all updates to the user object based on game logic, saving to db & returns the updated user
func CalculateUserUpdates(userData schema.User, wdb rdb.Database) (schema.User, error) {
	log.Debug.Println(log.Cyan("-- Begin CalculateUserUpdates --"))
	userData = CalculateManaRegen(userData)
	userData = CalculateTravelArrived(userData)
	userData, harvestErr := CalculateResourcesHarvested(userData, wdb)
	if harvestErr != nil {
		return userData, harvestErr
	}

	// Save changes to DB
	
	log.Debug.Println(log.Cyan("-- End CalculateUserUpdates --"))
	return userData, nil
}