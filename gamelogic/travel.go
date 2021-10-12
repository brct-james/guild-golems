// Package gamelogic provides functions for game logic
package gamelogic

import (
	"strings"
	"time"

	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/schema"
)

// Update whether golem arrived at destination, return the updated userData
func CalculateTravelArrived(userData schema.User) (schema.User) {
	log.Debug.Println(log.Cyan("-- Begin CalculateTravelArrived --"))
	log.Debug.Printf("golems: %v", userData.Golems)
	for i, golem := range userData.Golems {
		if strings.EqualFold(golem.Status, "traveling") {
			// Success, traveling, check if complete
			arrTime := time.Unix(golem.ArrivalTime, 0)
			now := time.Now()
			if arrTime.Before(now) {
				// Travel complete
				log.Debug.Printf("%v before %v, setting to idle", arrTime, now)
				userData.Golems[i].Status = "idle"
			}
		}
	}
	log.Debug.Println(log.Cyan("-- End CalculateTravelArrived --"))
	return userData
}