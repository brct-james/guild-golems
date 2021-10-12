// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

// Defines generic resource type
type Resource struct {
	Thing
	CapacityPerUnit float64 `json:"capacity_per_unit" binding:"required"`
}

// Defines resource in an inventory, used in udb, not json/wdb
type InventoryResource struct {
	Resource
	Quantity int `json:"quantity" binding:"required"`
}