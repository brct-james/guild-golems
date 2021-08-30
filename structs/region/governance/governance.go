package governance

type Governance struct {
	Type string `json:"type"`
	RulingFaction string `json:"ruling_faction"`
	NotableLaws []struct{} `json:"notable_laws"`
	Ideology Ideology `json:"ideology"`
	Economy Economy `json:"economy"`
	ConnectedRegions ConnectedRegions `json:"connected_regions"`
	Locales []string `json:"locales"`
}

// func New(gType string, ruling_faction string, notable_laws []struct{}, ideology Ideology, economy Economy, connected_regions ConnectedRegions, locales []string) *Governance {
// 	return &Governance{
// 		Type: gType,
// 		RulingFaction: ruling_faction,
// 		NotableLaws: notable_laws,
// 		Ideology: ideology,
// 		Economy: economy,
// 		ConnectedRegions: connected_regions,
// 		Locales: locales,
// 	}
// }