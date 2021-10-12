// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines harvestable resource node
type ResourceNode struct {
	Thing
	HarvestTime int `json:"harvest_time" binding:"required"`
	DropTables []DropTable `json:"drop_tables" binding:"required"`
}

// Defines droptables
type DropTable struct {
	ResourceSymbol string `json:"resource_symbol" binding:"required"`
	Rarity float64 `json:"rarity" binding:"required"`
	HarvestAmount int `json:"harvest_amount" binding:"required"`
}