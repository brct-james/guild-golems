// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import "strings"

// Defines a user which has Name, Symbol, Description
type Golem struct {
	HasSymbol
	Archetype string `json:"archetype" binding:"required"`
	LocationSymbol string `json:"location-symbol" binding:"required"` 
	Status string `json:"status" binding:"required"`
}

// golem archetypes and abbreviations map
type GolemArchetype struct {
	Name string `json:"name" binding:"required"`
	Abbreviation string `json:"abbreviation" binding:"required"`
}
var GolemArchetypes = map[string]GolemArchetype {
	"invoker": {Name:"Invoker", Abbreviation:"INV"},
	"harvester": {Name:"Harvester", Abbreviation:"HRV"},
	"courier": {Name:"Courier", Abbreviation:"COR"},
	"artisan": {Name:"Artisan", Abbreviation:"ART"},
	"merchant": {Name:"Merchant", Abbreviation:"MRC"},
}


// Defines the schema for EnergyDetails - a struct containing information on golem energy
// type EnergyDetails struct {
// 	Energy float64 `json:"energy" binding:"required"`
// 	EnergyCap float64 `json:"energy-cap" binding:"required"`
// 	EnergyRegen float64 `json:"energy-regen" binding:"required"`
// 	LastEnergyTick int64 `json:"last-energy-tick" binding:"required"`
// }

func NewGolem(symbol string, archetype string) Golem {
	return Golem{
		HasSymbol: HasSymbol{
			Symbol: symbol,
		},
		Archetype: archetype,
		LocationSymbol: "A-G",
		Status: "idle",
	}
}

func DoesGolemArchetypeMatch(golem Golem, archetype string) bool {
	return strings.EqualFold(golem.Archetype, archetype)
}

func FilterGolemListByArchetype(golems []Golem, archetype string) []Golem {
	filteredList := make([]Golem, 0)
	for _, golem := range golems {
		if DoesGolemArchetypeMatch(golem, archetype) {
			filteredList = append(filteredList, golem)
		}
	}
	return filteredList
}