// Package schema defines database and JSON schema as structs
package schema

// Defines a 'Thing' which has Name, Symbol, Description
type Thing struct {
	Name string `json:"name"`
	Symbol string `json:"symbol"`
	Description string `json:"description"`
}

type World struct {
	Thing
	Regions []Region `json:"regions"`
}