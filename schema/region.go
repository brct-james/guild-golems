// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

type Region struct {
	Thing
	BorderRegionSymbols []string `json:"border_region_symbols" binding:"required"`
	LocaleSymbols []string `json:"locale_symbols" binding:"required"`
}