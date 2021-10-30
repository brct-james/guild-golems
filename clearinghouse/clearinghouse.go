// Package clearinghouse provides functions for handling player orders and interacting with markets
package clearinghouse

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/brct-james/guild-golems/gamelogic"
	"github.com/brct-james/guild-golems/log"
	"github.com/brct-james/guild-golems/rdb"
	"github.com/brct-james/guild-golems/schema"
)

type QueuedOrder struct {
	ReferenceString string `json:"reference_string" binding:"required"`
	Order schema.Order `json:"order" binding:"required"`
}

var OrderQueue = make(chan QueuedOrder)

var UserDatabase rdb.Database
var WorldDatabase rdb.Database

var orderSpool = make(map[string]schema.Order)

// Spools order for handling by Clearinghouse, not executed till .Execute(referenceString) called, returns reference string: ORDER#XXXXXX
func Spool(order schema.Order, username string, merchantSymbol string) (string) {
	var refstr strings.Builder
	refstr.WriteString(fmt.Sprintf("ORDER#%d|%s:%s|%s%d%s@%d", time.Now().Unix(), username, merchantSymbol, order.Type, order.Quantity, order.ItemSymbol, order.TargetPrice))
	if order.ForceExecution {
		refstr.WriteString("*")
	}
	orderSpool[refstr.String()] = order
	return refstr.String()
}

// Executes order specified by given reference string, returns bool whether was able to start execution
func Execute(reference string) (bool) {
	if _, ok := orderSpool[reference]; !ok {
		return false // could not execute as order not in spool at reference
	}
	OrderQueue <- QueuedOrder{
		ReferenceString: reference,
		Order: orderSpool[reference],
	}
	return true
}

// Handles the execution of orders from the OrderQueue channel as a goroutine
func ProcessMarketOrder() {
	order := <- OrderQueue
	log.Debug.Printf("PROCESSING: %s", order.ReferenceString)
	usernameRE := regexp.MustCompile(`\|(.*?)\:`)
	golemSymbolRE := regexp.MustCompile(`\:(.*?)\|`)
	username := usernameRE.FindStringSubmatch(order.ReferenceString)[1]
	golemSymbol := golemSymbolRE.FindStringSubmatch(order.ReferenceString)[1]
	
	// Get user
	userData, userFound, getErr := schema.GetUserByUsernameFromDB(username, UserDatabase)
	if getErr != nil {
		log.Error.Printf("Warning! Could not get user by username from DB in ProcessMarketOrder for username %s", username)
		return
	}
	if !userFound {
		log.Error.Printf("Warning! User not found for username %s in ProcessMarketOrder", username)
		return
	}

	// Get market and check whether transaction still valid
	marketPath := fmt.Sprintf("[\"%s\"]", order.Order.MarketSymbol)
	market, marketGetErr := schema.Market_get_from_db(WorldDatabase, marketPath)
	if marketGetErr != nil {
		log.Error.Printf("Warning! Market not found for path %s from symbol %s", marketPath, order.Order.MarketSymbol)
		return
	}
	switch order.Order.Type {
	case "SELL":
		// Check price still meets criteria, unless forcing execution
		marketPrice := gamelogic.CalculateMarketPrice(market.Pricing[order.Order.ItemSymbol], market.Stock[order.Order.ItemSymbol])
		if order.Order.ForceExecution {
			// Forcing execution regardless of price
			sellAndSave(userData, market, order, golemSymbol, marketPrice)
			return
		}
		if order.Order.TargetPrice > marketPrice {
			// No longer valid price, cancel order
			log.Debug.Printf("Cancelling %s as order.Order.TargetPrice %d > marketPrice %d", order.ReferenceString, order.Order.TargetPrice, marketPrice)
			return
		}
		// Valid, Execute order
		sellAndSave(userData, market, order, golemSymbol, marketPrice)
	default:
		log.Error.Printf("Warning! Invalid order.Order.Type found in ProcessMarketOrder")
	}
}

func sellAndSave(userData schema.User, market schema.Market, order QueuedOrder, golemSymbol string, marketPrice int) () {
	// Update inventories with new contents
	// Remove from golem inventory
	userData.Inventories[golemSymbol].Contents[order.Order.ItemSymbol] = userData.Inventories[golemSymbol].Contents[order.Order.ItemSymbol] - order.Order.Quantity
	// delete symbol in contents if empty, then if contents empty delete entry in Inventories
	if userData.Inventories[golemSymbol].Contents[order.Order.ItemSymbol] == 0 {
		if len(userData.Inventories[golemSymbol].Contents) == 1 {
			delete(userData.Inventories, golemSymbol)
		}
		delete(userData.Inventories[golemSymbol].Contents, order.Order.ItemSymbol)
	}
	// Add to market stock
	market.Stock[order.Order.ItemSymbol] += order.Order.Quantity
	// Give user coins
	userData.Coins += uint64(marketPrice * order.Order.Quantity)
	// WONT WORK TILL CONVERT GOLEMS TO MAP RATHER THAN SLICE
	// Change golem status and statusdetail
	// userData.Golems[golemSymbol].Status = "idle"
	// userData.Golems[golemSymbol].StatusDetail = ""

	// Save both to db
	// Exclusively update the specific inventory path, coin path, and golem path in the user data to avoid race conditions
	// golemsPath := fmt.Sprintf(".golems.[\"%s\"]", golemSymbol)
	inventoryPath := fmt.Sprintf("[\"inventories.%s\"]", golemSymbol)
	saveCoinsErr := schema.SaveUserDataAtPathToDB(UserDatabase, userData, ".coins", userData.Coins)
	if saveCoinsErr != nil {
		log.Error.Printf("Warning! Error while saving after SELL market order: %v", saveCoinsErr)
	}
	// saveGolemsErr := schema.SaveUserDataAtPathToDB(UserDatabase, userData, golemsPath, userData.Coins)
	// if saveGolemsErr != nil {
	// 	log.Error.Printf("Warning! Error while saving after SELL market order: %v", saveGolemsErr)
	// }
	saveInventoryErr := schema.SaveUserDataAtPathToDB(UserDatabase, userData, inventoryPath, userData.Coins)
	if saveInventoryErr != nil {
		log.Error.Printf("Warning! Error while saving after SELL market order: %v", saveInventoryErr)
	}
	saveMarketErr := schema.Market_save_to_db(WorldDatabase, market)
	if saveMarketErr != nil {
		log.Error.Printf("Warning! Error while saving after SELL market order: %v", saveMarketErr)
	}
}