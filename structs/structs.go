package structs

import g "github.com/brct-james/guild-golems/structs/region/governance"

type Region struct {
	Name string `json:"name"`
	Symbol string `json:"symbol"`
	Description string `json:"description"`
	CapitalCity string `json:"capital_city"`
	Governance g.Governance `json:"governance"`
}

func New(name string, symbol string, description string, capital string, governance g.Governance) *Region {
	return &Region{
		Name: name,
		Symbol: symbol,
		Description: description,
		CapitalCity: capital,
		Governance: governance,
	}
}