// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

import (
	"encoding/json"

	"github.com/brct-james/guild-golems/log"
)

// Defines generic recipe type
type Recipe struct {
	Thing
	ProductAmount int `json:"product_amount" binding:"required"`
	CraftDuration int `json:"craft_duration" binding:"required"`
	Incredients map[string]int `json:"ingredients" binding:"required"`
}

var Recipes map[string]Recipe

// Unmarshals recipe from json byte array
func Recipe_unmarshal_json(recipe_json []byte) (Recipe, error) {
	log.Debug.Println("Unmarshalling recipe.json")
	var recipe Recipe
	err := json.Unmarshal(recipe_json, &recipe)
	if err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}

// Unmarshals all recipes from json byte array
func Recipe_unmarshal_all_json(recipe_json []byte) (map[string]Recipe, error) {
	log.Debug.Println("Unmarshalling recipe.json")
	nilRecipe := make(map[string]Recipe)
	var recipes map[string]Recipe
	err := json.Unmarshal(recipe_json, &recipes)
	if err != nil {
		return nilRecipe, err
	}
	return recipes, nil
}