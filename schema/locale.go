// Package schema defines database and JSON schema as structs
package schema

// Defines the characteristics of Locales (e.g. cities, forests, etc.)
type Locale struct {
	Thing
	ResourceNodes []ResourceNode `json:"resource_nodes"`
	Routes []Route `json:"routes"`
}

// Defines the characteristics of Routes between locales
type Route struct {
	Thing
	DangerLevel int `json:"danger_level"`
	TravelTime int `json:"travel_time"`
}