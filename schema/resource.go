// Package schema defines database and JSON schema as structs
package schema

// Defines generic resource type
type Resource struct {
	Thing
	Quantity int `json:"quantity"`
}

// Defines harvestable resource node
type ResourceNode struct {
	Resource
	MaxQuantity int `json:"max_quantity"`
	RenewalRate int `json:"renewal_rate"`
	HarvestTime int `json:"harvest_time"`
	DropTable []HarvestableResource `json:"drop_table"`
}

// Defines harvestable resource
type HarvestableResource struct {
	Resource
	Rarity float64 `json:"rarity"`
}

// Defines inventory resource
type InventoryResource struct {
	Resource
	LocationSymbol string `json:"location_symbol"`
}