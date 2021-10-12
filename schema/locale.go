// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines the characteristics of Locales (e.g. cities, forests, etc.)
type Locale struct {
	Thing
	ResourceNodes []ResourceNode `json:"resource_nodes" binding:"required"`
	Routes []Route `json:"routes" binding:"required"`
}

// Defines the characteristics of Routes between locales
// TravelTime in seconds
type Route struct {
	Thing
	DangerLevel int `json:"danger_level" binding:"required"`
	TravelTime int `json:"travel_time" binding:"required"`
	Cost int `json:"cost" binding:"required"`
}