// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines Ritual struct
type Ritual struct {
	Thing
	ManaCost float64 `json:"mana-cost" binding:"required"`
}

func NewRitual(name string, symbol string, description string, manaCost float64) Ritual {
	return Ritual{
		Thing: Thing{
			Name: name,
			HasSymbol: HasSymbol{
				Symbol: symbol,
			},
			Description: description,
		},
		ManaCost: manaCost,
	}
}