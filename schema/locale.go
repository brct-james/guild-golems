// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines the characteristics of Locales (e.g. cities, forests, etc.)
type Locale struct {
	Thing
	ResourceNodeSymbols []string `json:"resource_node_symbols" binding:"required"`
	RouteSymbols []string `json:"route_symbols" binding:"required"`
}