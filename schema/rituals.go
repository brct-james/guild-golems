// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines Ritual struct
type Ritual struct {
	Thing
	ManaCost float64 `json:"mana-cost" binding:"required"`
}

// ritual info map
var Rituals = map[string]Ritual {
	"summon-invoker": NewRitual("Summon Invoker", "summon-invoker", "Spend mana to summon a new invoker, who can be used to help generate even more mana.", 600),
	"summon-harvester": NewRitual("Summon Harvester", "summon-harvester", "Spend mana to summon a new harvester, who can be used to gather resources from nodes in the world.", 600),
	"summon-courier": NewRitual("Summon Courier", "summon-courier", "Spend mana to summon a new courier, who can be used to transport resources between locales", 600),
	"summon-merchant": NewRitual("Summon Merchant", "summon-merchant", "Spend mana to summon a new merchant, who can be used to buy and sell resources at markets", 600),
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