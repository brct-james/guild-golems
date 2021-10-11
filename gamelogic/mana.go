// Package gamelogic provides functions for game logic
package gamelogic

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/brct-james/guild-golems/responses"
	"github.com/brct-james/guild-golems/schema"
)

// Attempt to do something using mana, return if successful and the new mana value
func TryManaPurchase(w http.ResponseWriter, mana float64, manaCost float64) (bool, float64) {
	if mana < manaCost {
		// fail state, not enough mana
		lowManaMsg := fmt.Sprintf("Have %v but Requires %v", mana, manaCost)
		responses.SendRes(w, responses.Not_Enough_Mana, nil, lowManaMsg)
		return false, mana
	}
	return true, mana-manaCost
}

// Update mana value based on time since last update, return the updated userData
func CalculateManaRegen(userData schema.User) (schema.User) {
	secondsSinceTick := time.Since(time.Unix(userData.LastManaTick, 0)).Seconds()
	numInvokers := len(schema.FilterGolemListByArchetype(userData.Golems, "invoker"))
	userData.Mana = math.Min(userData.ManaCap, userData.Mana + (secondsSinceTick * (userData.ManaRegen + (float64(numInvokers)/2))))
	userData.LastManaTick = time.Now().Unix()
	return userData
}