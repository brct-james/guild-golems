// Package gamelogic provides functions for game logic
package gamelogic

import (
	"fmt"
	"net/http"

	"github.com/brct-james/guild-golems/responses"
)

func TryManaPurchase(w http.ResponseWriter, mana float64, manaCost float64) (bool, float64) {
	if mana < manaCost {
		// fail state, not enough mana
		lowManaMsg := fmt.Sprintf("Have %v but Requires %v", mana, manaCost)
		responses.SendRes(w, responses.Not_Enough_Mana, nil, lowManaMsg)
		return false, mana
	}
	return true, mana-manaCost
}