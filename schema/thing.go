// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines HasSymbol

type HasSymbol struct {
	Symbol string `json:"symbol" binding:"required"`
}

// Defines a 'Thing' which has Name, Symbol, Description
type Thing struct {
	HasSymbol
	Name string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}