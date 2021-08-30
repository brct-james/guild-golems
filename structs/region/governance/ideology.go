package governance

type Ideology struct {
	EconomicFreedom EconomicFreedom `json:"economic_freedom"`
	SocialValues SocialValues `json:"social_values"`
	StateAuthority StateAuthority	`json:"state_authority"`
}

type EconomicFreedom int

const (
	Unregulated = iota
	Semi_Regulated
	Regulated
	Semi_Planned
	Planned
)

type SocialValues int

const (
	Collectivist = iota
	Progressive
	Centrist
	Conservative
	Individualist
)

type StateAuthority int

const (
	Anarchic = iota
	Libertarian
	Moderate
	Controlling
	Authoritarian
)