// Package gamelogic provides functions for game logic
package gamelogic

import (
	"strings"
	"time"

	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/schema"
	"github.com/brct-james/guild-golems/timecalc"
)

// Update whether golem arrived at destination, return the updated userData
func CalculateTravelArrived(userData schema.User) (schema.User) {
	log.Debug.Println(log.Cyan("-- Begin CalculateTravelArrived --"))
	log.Debug.Printf("golems: %v", userData.Golems)
	for i, golem := range userData.Golems {
		if strings.EqualFold(golem.Status, "traveling") {
			// Success, traveling, check if complete
			arrTime := time.Unix(golem.TravelInfo.ArrivalTime, 0)
			now := time.Now()
			if arrTime.Before(now) {
				// Travel complete
				log.Debug.Printf("%v before %v, setting to idle", arrTime, now)
				userData.Golems[i].Status = "idle"
				userData.Golems[i].LocationSymbol = userData.Golems[i].StatusDetail
				userData.Golems[i].StatusDetail = ""
			}
		}
	}
	log.Debug.Println(log.Cyan("-- End CalculateTravelArrived --"))
	return userData
}

func CalcualteArrivalTime(travelTime int, archetype string) (time.Time) {
	// Buff travel time for couriers
	if archetype == "courier" {
		travelTime = int(float64(travelTime) * 0.75)
	}
	return timecalc.AddSecondsToTimestamp(time.Now(), travelTime)

}