package governance

type ConnectedRegions struct {
	Symbol string `json:"symbol"`
	TravelTime int64 `json:"travel_time"`
	RouteDanger RouteDanger `json:"route_danger"`
}

type RouteDanger int

const (
	None = iota
	Low
	Some
	High
	Suicidal
)