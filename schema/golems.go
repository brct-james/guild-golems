// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"errors"
	"fmt"
	"strings"
)

// Defines a user which has Name, Symbol, Description
type Golem struct {
	HasSymbol
	Archetype string `json:"archetype" binding:"required"`
	LocationSymbol string `json:"location_symbol" binding:"required"` 
	Status string `json:"status" binding:"required"`
	StatusDetail string `json:"status-detail" binding:"required"`
	Capacity float64 `json:"capacity" binding:"required"`
	Inventory Inventory `json:"inventory" binding:"required"`
	TravelInfo GolemTravelInfo `json:"travel_info" binding:"required"`
}

func UpdateGolemLinkedData(userData User, targetGolem Golem) (Golem) {
	targetGolem.Inventory = userData.Inventories[targetGolem.Symbol]
	return targetGolem
}

func UpdateGolemListLinkedData(userData User, golemList []Golem) ([]Golem) {
	for k, g := range golemList {
		g.Inventory = userData.Inventories[g.Symbol]
		golemList[k] = g
	}
	return golemList
}

// Defines relevant info for golems while traveling
type GolemTravelInfo struct {
	ArrivalTime int64 `json:"arrival_time" binding:"required"`
	OriginSymbol string `json:"origin_symbol" binding:"required"`
	DestinationSymbol string `json:"destination_symbol" binding:"required"`
	RouteDanger int `json:"route_danger" binding:"required"`
}

// golem statuses map
type GolemStatus struct {
	Name string `json:"name" binding:"required"`
	IsBlocking bool `json:"is-blocking" binding:"required"`
}
var GolemStatuses = map[string]GolemStatus {
	"idle": {Name:"Idle", IsBlocking: false},
	"harvesting": {Name:"Harvesting", IsBlocking: false},
	"traveling": {Name:"Traveling", IsBlocking: true},
	"invoking": {Name:"Invoking", IsBlocking: true},
	"packing": {Name:"Packing", IsBlocking: true},
	"storing": {Name:"Storing", IsBlocking: true},
}

// golem archetypes and abbreviations map
type GolemArchetype struct {
	Name string `json:"name" binding:"required"`
	Abbreviation string `json:"abbreviation" binding:"required"`
	AllowedStatuses []string `json:"allowed-statuses" binding:"required"`
}
var GolemArchetypes = map[string]GolemArchetype {
	"invoker": {Name:"Invoker", Abbreviation:"INV",
		AllowedStatuses: []string{"idle", "invoking"},
	},
	"harvester": {Name:"Harvester", Abbreviation:"HRV",
		AllowedStatuses: []string{"idle", "traveling", "harvesting"},
	},
	"courier": {Name:"Courier", Abbreviation:"COR",
		AllowedStatuses: []string{"idle", "traveling", "packing", "storing"},
	},
	"artisan": {Name:"Artisan", Abbreviation:"ART",
		AllowedStatuses: []string{"idle", "traveling"},
	},
	"merchant": {Name:"Merchant", Abbreviation:"MRC",
		AllowedStatuses: []string{"idle", "traveling"},
	},
}

// Defines the structure for golem status update requests
// Instructions expects an object/map with different keys depending on newStatus
type GolemStatusUpdateBody struct {
	NewStatus string `json:"new_status" binding:"required"`
	Instructions interface{} `json:"instructions" binding:"required"`
}


// Defines the schema for EnergyDetails - a struct containing information on golem energy
// type EnergyDetails struct {
// 	Energy float64 `json:"energy" binding:"required"`
// 	EnergyCap float64 `json:"energy-cap" binding:"required"`
// 	EnergyRegen float64 `json:"energy-regen" binding:"required"`
// 	LastEnergyTick int64 `json:"last-energy-tick" binding:"required"`
// }

func NewGolem(symbol string, archetype string, location string, startingStatus string, capacity float64) Golem {
	return Golem{
		HasSymbol: HasSymbol{
			Symbol: symbol,
		},
		Archetype: archetype,
		LocationSymbol: location,
		Status: startingStatus,
		StatusDetail: "",
		Capacity: capacity,
		Inventory: Inventory{
			LocationSymbol: symbol,
			Contents: make(map[string]int),
		},
		TravelInfo: GolemTravelInfo{
			ArrivalTime: 0,
			OriginSymbol: "",
			DestinationSymbol: "",
			RouteDanger: 0,
		},
	}
}

func IsStatusAllowedForArchetype(archetype string, newStatus string) (bool, error) {
	archetypeInfo, ok := GolemArchetypes[archetype]
	if !ok {
		// Fail case - golem archetype not in list of valid statuses
		resMsg1 := fmt.Sprintf("Specified archetype %s not in list of valid archetypes", archetype)
		return false, errors.New(resMsg1)
	}
	for _, status := range archetypeInfo.AllowedStatuses {
		if strings.EqualFold(status, newStatus) {
			// Success case, archetype is allowed
			return true, nil
		}
	}
	// Fail case, archetype not allowed
	return false, nil
}

func DoesGolemArchetypeMatch(golem Golem, archetype string) bool {
	return strings.EqualFold(golem.Archetype, archetype)
}

func DoesGolemStatusMatch(golem Golem, status string) bool {
	return strings.EqualFold(golem.Status, status)
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

func FilterGolemListByStatus(golems []Golem, status string) []Golem {
	filteredList := make([]Golem, 0)
	for _, golem := range golems {
		if DoesGolemStatusMatch(golem, status) {
			filteredList = append(filteredList, golem)
		}
	}
	return filteredList
}

func FindIndexOfGolemWithSymbol(golems []Golem, symbol string) (bool, int) {
	for i := range golems {
		if strings.EqualFold(golems[i].Symbol, symbol) {
			return true, i
		}
	}
	return false, -1
}