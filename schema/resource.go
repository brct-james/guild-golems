// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines generic resource type
type Resource struct {
	Thing
	Quantity int `json:"quantity" binding:"required"`
}

// Defines harvestable resource node
type ResourceNode struct {
	Resource
	MaxQuantity int `json:"max_quantity" binding:"required"`
	RenewalRate int `json:"renewal_rate" binding:"required"`
	HarvestTime int `json:"harvest_time" binding:"required"`
	DropTable []HarvestableResource `json:"drop_table" binding:"required"`
}

// Defines harvestable resource
type HarvestableResource struct {
	Resource
	Rarity float64 `json:"rarity" binding:"required"`
}

// Defines inventory resource
type InventoryResource struct {
	Resource
	LocationSymbol string `json:"location_symbol" binding:"required"`
}