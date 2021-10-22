// Package gamelogic provides functions for game logic
package gamelogic

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/schema"
)

// Update inventory based on time since last harvest, return the updated userData
func CalculateResourcesHarvested(userData schema.User, wdb rdb.Database) (schema.User, error) {
	log.Debug.Println(log.Cyan("-- Begin CalculateResourcesHarvested --"))
	if len(userData.Golems) < 1 {
		// no golems
		return userData, nil
	}
	harvesters := schema.FilterGolemListByArchetype(userData.Golems, "harvester")
	if len(harvesters) < 1 {
		// no harvesters
		return userData, nil
	}
	relevantGolems := schema.FilterGolemListByStatus(harvesters,"harvesting")
	if len(relevantGolems) < 1 {
		// no harvesting harvesters
		return userData, nil
	}

	secondsSinceTick := time.Since(time.Unix(userData.LastHarvestTick, 0)).Seconds()
	for _, golem := range relevantGolems {
		if golem.StatusDetail == "" {
			log.Error.Printf("")
		}
		resNodePath := fmt.Sprintf("[\"%s\"]", golem.StatusDetail)
		node, resNodeErr := schema.ResourceNode_get_from_db(wdb, resNodePath)
		if resNodeErr != nil {
			resNodeErrMsg := fmt.Sprintf("User %s with Golem %s has StatusDetail %s but encountered error attempting to get ResourceNode at that path from db: %v", userData.Username, golem.Symbol, golem.StatusDetail, resNodeErr)
			log.Error.Printf(resNodeErrMsg)
			return userData, resNodeErr
		}
		remainingSeconds := int(secondsSinceTick) % node.HarvestTime
		numHarvestsSinceTick := (int(secondsSinceTick)-remainingSeconds) / node.HarvestTime
		for _, drop := range node.DropTables {
			// Calculate whether drops
			if drop.Rarity < 1 {
				if rand.Float64() < drop.Rarity {
					continue
				}
			}
			// Create locationInventory for this symbol if not already exists
			locationInventory, ok := userData.Inventory[golem.LocationSymbol]
			if !ok {
				userData.Inventory[golem.LocationSymbol] = schema.LocationInventory{
					LocationSymbol: golem.LocationSymbol,
					Contents: make(map[string]int),
				}
				locationInventory = userData.Inventory[golem.LocationSymbol]
			}
			// Add drop amount to inventory
			// TODO: Check for max capacity by looking up InventoryResource definition in db
			userData.Inventory[golem.LocationSymbol].Contents[drop.ResourceSymbol] = locationInventory.Contents[drop.ResourceSymbol] + (drop.HarvestAmount * numHarvestsSinceTick)
		}
		userData.LastHarvestTick = time.Now().Unix()
	}
	log.Debug.Println(log.Cyan("-- End CalculateResourcesHarvested --"))
	return userData, nil
}

// // Update mana value based on time since last update, return the updated userData
// func CalculateManaRegen(userData schema.User) (schema.User) {
// 	log.Debug.Println(log.Cyan("-- Begin CalculateManaRegen --"))
// 	secondsSinceTick := time.Since(time.Unix(userData.LastManaTick, 0)).Seconds()
// 	numInvokers := len(schema.FilterGolemListByStatus(schema.FilterGolemListByArchetype(userData.Golems, "invoker"), "invoking"))
// 	userData.Mana = math.Min(userData.ManaCap, userData.Mana + (secondsSinceTick * (userData.ManaRegen + (float64(numInvokers)/2))))
// 	userData.LastManaTick = time.Now().Unix()
// 	log.Debug.Println(log.Cyan("-- End CalculateManaRegen --"))
// 	return userData
// }

// // Update whether golem arrived at destination, return the updated userData
// func CalculateTravelArrived(userData schema.User) (schema.User) {
// 	log.Debug.Println(log.Cyan("-- Begin CalculateTravelArrived --"))
// 	log.Debug.Printf("golems: %v", userData.Golems)
// 	for i, golem := range userData.Golems {
// 		if strings.EqualFold(golem.Status, "traveling") {
// 			// Success, traveling, check if complete
// 			arrTime := time.Unix(golem.TravelInfo.ArrivalTime, 0)
// 			now := time.Now()
// 			if arrTime.Before(now) {
// 				// Travel complete
// 				log.Debug.Printf("%v before %v, setting to idle", arrTime, now)
// 				userData.Golems[i].Status = "idle"
// 			}
// 		}
// 	}
// 	log.Debug.Println(log.Cyan("-- End CalculateTravelArrived --"))
// 	return userData
// }