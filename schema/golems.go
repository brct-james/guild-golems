// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines a user which has Name, Symbol, Description
type Golem struct {
	HasSymbol
	Purpose string `json:"purpose" binding:"required"`
	LocationSymbol string `json:"location-symbol" binding:"required"` 
	Status string `json:"status" binding:"required"`
}

// Defines the schema for EnergyDetails - a struct containing information on golem energy
// type EnergyDetails struct {
// 	Energy float64 `json:"energy" binding:"required"`
// 	EnergyCap float64 `json:"energy-cap" binding:"required"`
// 	EnergyRegen float64 `json:"energy-regen" binding:"required"`
// 	LastEnergyTick int64 `json:"last-energy-tick" binding:"required"`
// }

func NewGolem(symbol string, purpose string) Golem {
	return Golem{
		HasSymbol: HasSymbol{
			Symbol: symbol,
		},
		Purpose: purpose,
		LocationSymbol: "A-G",
		Status: "idle",
	}
}

func DoesGolemPurposeMatch(golem Golem, purpose string) bool {
	return golem.Purpose == purpose
}

func FilterGolemListByPurpose(golems []Golem, purpose string) []Golem {
	filteredList := make([]Golem, 0)
	for _, golem := range golems {
		if DoesGolemPurposeMatch(golem, purpose) {
			filteredList = append(filteredList, golem)
		}
	}
	return filteredList
}