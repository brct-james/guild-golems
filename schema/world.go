// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

type World struct {
	Thing
	RegionSymbols []string `json:"region_symbols" binding:"required"`
}