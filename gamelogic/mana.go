// Package gamelogic provides functions for game logic
package gamelogic

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/brct-james/guild-golems/gamevars"
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/responses"
	"github.com/brct-james/guild-golems/schema"
)

// Attempt to do something using mana, return if successful and the new mana value
func TryManaPurchase(w http.ResponseWriter, mana float64, manaCost float64) (bool, float64) {
	log.Debug.Println(log.Cyan("-- Begin TryManaPurchase --"))
	if mana < manaCost {
		// fail state, not enough mana
		lowManaMsg := fmt.Sprintf("Have %v but Requires %v", mana, manaCost)
		responses.SendRes(w, responses.Not_Enough_Mana, nil, lowManaMsg)
		log.Debug.Println(log.Cyan("-- End TryManaPurchase --"))
		return false, mana
	}
	log.Debug.Println(log.Cyan("-- End TryManaPurchase --"))
	return true, mana-manaCost
}

// Update mana value based on time since last update, return the updated userData
func CalculateManaRegen(userData schema.User) (schema.User) {
	log.Debug.Println(log.Cyan("-- Begin CalculateManaRegen --"))
	secondsSinceTick := time.Since(time.Unix(userData.LastManaTick, 0)).Seconds()
	numInvokers := len(schema.FilterGolemMapByStatus(userData.Golems, "invoking"))
	userData.Mana = math.Min(userData.ManaCap, userData.Mana + (secondsSinceTick * (userData.ManaRegen + (float64(numInvokers)*gamevars.Invoker_Potency))))
	userData.LastManaTick = time.Now().Unix()
	log.Debug.Println(log.Cyan("-- End CalculateManaRegen --"))
	return userData
}