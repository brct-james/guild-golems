package governance

type Economy struct {
	RelativePopulation RelativePopulation `json:"relative_population"`
	RelativeSize RelativeSize `json:"relative_size"`
	RelativeWealth RelativeWealth `json:"relative_wealth"`
	RegionTaxRate float32 `json:"region_tax_rate"`
}

type RelativePopulation int

const (
	Uninhabited = iota
	Barely_Inhabited
	Settlement
	Town
	City
	Megalopolis
)

type RelativeSize int

const (
	Miniscule = iota
	Tiny
	Small
	Medium
	Large
	Extreme
)

type RelativeWealth int

const (
	Destitute = iota
	Impoverished
	Liveable
	Middle_Class
	Well_Off
	Wealthy
	One_Percent
)