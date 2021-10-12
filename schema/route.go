// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines the characteristics of Routes between locales
// TravelTime in seconds
type Route struct {
	Thing
	DangerLevel int `json:"danger_level" binding:"required"`
	TravelTime int `json:"travel_time" binding:"required"`
	Cost int `json:"cost" binding:"required"`
}