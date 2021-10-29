// Package gamevars provides initial & meta values for gamelogic
package gamevars

// Golem Capacity
var Capacity_Invoker float64 = 0
var Capacity_Harvester float64 = 10
var Capacity_Courier float64 = 100

// Base Locale Capacity
// UNIMPLEMENTED
var Base_Locale_Capacity float64 = 100

// Invoker Potency
var Invoker_Potency float64 = 0.5

// Starting Mana
var Starting_Mana_Regen float64 = 1
var Starting_Mana_Cap float64 = 21600
var Starting_Mana float64 = 3600

// Starting Coins
var Starting_Coins uint64 = 0

// Starting Rituals
var Starting_Rituals []string = []string{
	"summon-invoker",
	"summon-harvester",
	"summon-courier",
}

// Starting Location
var Starting_Location = "A-G"

// Market Consumption Rate
var Market_Consumption_Rate int64 = 60