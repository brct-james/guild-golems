package governance

type Governance struct {
	Type string `json:"type"`
	RulingFaction string `json:"ruling_faction"`
	NotableLaws []struct{} `json:"notable_laws"`
	Ideology Ideology `json:"ideology"`
}

// func New(gType string, ruling_faction string, notable_laws []struct{}, ideology Ideology) *Governance {
// 	return &Governance{
// 		Type: gType,
// 		RulingFaction: ruling_faction,
// 		NotableLaws: notable_laws,
// 		Ideology: ideology,
// 	}
// }