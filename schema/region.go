// Package schema defines database and JSON schema as structs
package schema

type Region struct {
	Thing
	BorderRegionSymbols []string `json:"border_region_symbols"`
	Locales []Locale `json:"locales"`
}