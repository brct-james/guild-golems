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

type MarketOrder struct {
	ReferenceString string `json:"reference_string" binding:"required"`
	Status string `json:"status" binding:"required"`
	Order schema.Order `json:"order" binding:"required"`
}

var OrderQueue = make(chan string)
var Orders = make(map[string]MarketOrder)

var UserDatabase rdb.Database
var WorldDatabase rdb.Database

var lastNano = time.Now().Unix()
var nanoMod = 0

// Slice of valid order statuses
var OrderStatuses = []string{"Spooled", "Queued", "InProcessing", "InCancelling", "Cancelled", "InExecuting", "Executed"}

// Gets orders by user, status="*" ignores status
func GetOrdersByUserWithStatus(username string, status string) ([]MarketOrder) {
	ignoreStatus := false
	if status == "*" {
		ignoreStatus = true
	}
	res := make([]MarketOrder, 0)
	for key, order := range Orders {
		usernameRE := regexp.MustCompile(`\|(.*?)\:`)
		orderUser := usernameRE.FindStringSubmatch(key)[1]
		if orderUser == username && (ignoreStatus || strings.EqualFold(order.Status, status)) {
			res = append(res, order)
		}
	}
	return res
}

// Spools order for handling by Clearinghouse, not executed till .Execute(referenceString) called, returns reference string: ORDER#XXXXXX
func Spool(order schema.Order, username string, merchantSymbol string) (string) {
	log.Debug.Println(log.Yellow("-- Clearinghouse:Spool --"))
	var refstr strings.Builder
	
	// Calc nanoMod if multiple orders this second
	newNano := time.Now().Unix()
	if lastNano == newNano {
		nanoMod++
	} else {
		nanoMod = 0
	}
	// Create reference string
	refstr.WriteString(fmt.Sprintf("ORDER#%d+%d|%s:%s|%s%d%s@%d", newNano, nanoMod, username, merchantSymbol, order.Type, order.Quantity, order.ItemSymbol, order.TargetPrice))
	lastNano = newNano
	if order.ForceExecution {
		refstr.WriteString("*")
	}

	Orders[refstr.String()] = MarketOrder{
		ReferenceString: refstr.String(),
		Status: "Spooled",
		Order: order,
	}
	log.Debug.Println(log.Yellow("-- End Clearinghouse:Spool --"))
	return refstr.String()
}

// Executes order specified by given reference string, returns bool whether was able to start execution
func Execute(reference string) (bool) {
	log.Debug.Println(log.Yellow("-- Clearinghouse:Execute --"))
	order, ok := Orders[reference]
	if !ok {
		log.Error.Printf("Error: reference string %s not in Orders", reference)
		log.Debug.Println(log.Yellow("-- End Clearinghouse:Execute --"))
		return false // could not execute as order not in spool at reference
	}
	log.Debug.Println("Queueing order in OrderQueue channel")
	changeOrderStatus(order, "Queued")
	OrderQueue <- reference
	log.Debug.Println(log.Yellow("-- End Clearinghouse:Execute --"))
	return true
}

func getUser(username string, rid string, order MarketOrder, golemSymbol string) (bool, schema.User) {
	// Get user
	userData, userFound, getErr := schema.GetUserByUsernameFromDB(username, UserDatabase)
	if getErr != nil {
		log.Error.Printf("[%s] Warning! Could not get user by username from DB in ProcessMarketOrder for username %s", rid, username)
		cancelOrder(rid, order, userData, golemSymbol)
		log.RoutineDebug.Printf(log.Yellow(fmt.Sprintf("-- End Clearinghouse:ProcessMarketOrder | %s --", rid)))
		return false, userData
	}
	if !userFound {
		log.Error.Printf("[%s] Warning! User not found for username %s in ProcessMarketOrder", rid, username)
		cancelOrder(rid, order, userData, golemSymbol)
		log.RoutineDebug.Printf(log.Yellow(fmt.Sprintf("-- End Clearinghouse:ProcessMarketOrder | %s --", rid)))
		return false, userData
	}
	return true, userData
}

func getMarket(userData schema.User, rid string, order MarketOrder, golemSymbol string) (bool, schema.Market) {
	// Get market and check whether transaction still valid
	marketPath := fmt.Sprintf("[\"%s\"]", order.Order.MarketSymbol)
	market, marketGetErr := schema.Market_get_from_db(WorldDatabase, marketPath)
	if marketGetErr != nil {
		log.Error.Printf("[%s] Warning! Market not found for path %s from symbol %s", rid, marketPath, order.Order.MarketSymbol)
		cancelOrder(rid, order, userData, golemSymbol)
		log.RoutineDebug.Printf(log.Yellow(fmt.Sprintf("-- End Clearinghouse:ProcessMarketOrder | %s --", rid)))
		return false, market
	}
	return true, market
}

// Handles the execution of orders from the OrderQueue channel as a goroutine
func ProcessMarketOrder(args ...interface{}) {
	rid := <- OrderQueue // Routine ID / Reference String
	order := Orders[rid]
	changeOrderStatus(order, "InProcessing")
	log.RoutineDebug.Printf(log.Yellow(fmt.Sprintf("-- Clearinghouse:ProcessMarketOrder | %s --", rid)))
	usernameRE := regexp.MustCompile(`\|(.*?)\:`)
	golemSymbolRE := regexp.MustCompile(`\:(.*?)\|`)
	username := usernameRE.FindStringSubmatch(rid)[1]
	golemSymbol := golemSymbolRE.FindStringSubmatch(rid)[1]
	
	gotUser, userData := getUser(username, rid, order, golemSymbol)
	if !gotUser { return } // Handled by getuser

	gotMarket, market := getMarket(userData, rid, order, golemSymbol)
	if !gotMarket { return } // Handled by getmarket
		
	// Calculate Market Price
	marketPrice := gamelogic.CalculateMarketPrice(market.Pricing[order.Order.ItemSymbol], market.Stock[order.Order.ItemSymbol])
	log.RoutineDebug.Printf("[%s] TargetPrice: %d | MarketPrice: %d | ForceExecution? %t", rid, order.Order.TargetPrice, marketPrice, order.Order.ForceExecution)

	// Handle order
	log.RoutineDebug.Printf("[%s] %s ORDER", rid, order.Order.Type)
	switch order.Order.Type {
	case "SELL":
		// Sell conditions met?
		// Yes if Forcing Execution
		if order.Order.ForceExecution {
			// Forcing execution regardless of price
			executeSellOrder(rid, userData, market, order, golemSymbol, marketPrice)
			log.RoutineDebug.Printf(log.Yellow(fmt.Sprintf("-- End Clearinghouse:ProcessMarketOrder | %s --", rid)))
			return
		}

		// Yes if prices still meet criteria - else cancel
		if order.Order.TargetPrice > marketPrice {
			// No longer valid price, cancel order
			log.RoutineDebug.Printf("[%s] Cancelling as order.Order.TargetPrice %d > marketPrice %d", rid, order.Order.TargetPrice, marketPrice)
			cancelOrder(rid, order, userData, golemSymbol)
			log.RoutineDebug.Printf(log.Yellow(fmt.Sprintf("-- End Clearinghouse:ProcessMarketOrder | %s --", rid)))
			return
		}

		// ExecuteSellOrder
		log.RoutineDebug.Printf("[%s] execute sell order", rid)
		// Valid, Execute order
		executeSellOrder(rid, userData, market, order, golemSymbol, marketPrice)
	default:
		log.Error.Printf("[%s] Warning! Invalid order.Order.Type found in ProcessMarketOrder", rid)
		cancelOrder(rid, order, userData, golemSymbol)
		return
	}
	log.RoutineDebug.Printf(log.Yellow(fmt.Sprintf("-- End Clearinghouse:ProcessMarketOrder | %s --", rid)))
}

func changeOrderStatus(order MarketOrder, newStatus string) (MarketOrder) {
	log.RoutineDebug.Printf("[%s] %s", order.ReferenceString, log.Bold(fmt.Sprintf("%s ---> %s", order.Status, newStatus)))
	order.Status = newStatus
	Orders[order.ReferenceString] = order
	return order
}

func getInventoryPath(golemSymbol string) string {
	return fmt.Sprintf("[\"inventories\"][\"%s\"]", golemSymbol)
}

func deleteGolemInventory(token string, ip string) {
	numPathsDeleted, delErr := UserDatabase.DelJsonData(token, ip)
	if delErr != nil {
		log.Error.Printf("Warning! Error while deleting golem inventory with path: %s. Err: %v", ip, delErr)
	}
	log.Debug.Printf("Using %s, %d paths deleted", ip, numPathsDeleted)
}

func updateGolemInventory(token string, ip string, inv schema.Inventory) {
	saveInventoryErr := schema.SaveUserDataAtPathToDB(UserDatabase, token, ip, inv)
	if saveInventoryErr != nil {
		log.Error.Printf("Warning! Error while saving after SELL market order: %v", saveInventoryErr)
	}
}

func cancelOrder(rid string, order MarketOrder, userData schema.User, golemSymbol string) {
	changeOrderStatus(order, "InCancelling")
	golem, ok := userData.Golems[golemSymbol]
	if !ok {
		log.Error.Printf("[%s] Warning! Error while cancelling market order!", rid)
	}
	golem.Status = "idle"
	golem.StatusDetail = ""
	saveGolem(rid, userData.Token, golem)
	changeOrderStatus(order, "Cancelled")
}

func saveGolem(rid string, token string, golem schema.Golem) {
	golemsPath := fmt.Sprintf("[\"golems\"][\"%s\"]", golem.Symbol)
	saveGolemsErr := schema.SaveUserDataAtPathToDB(UserDatabase, token, golemsPath, golem)
	if saveGolemsErr != nil {
		log.Error.Printf("[%s] Warning! Error while saving after SELL market order: %v", rid, saveGolemsErr)
	}
}

func executeSellOrder(rid string, userData schema.User, market schema.Market, order MarketOrder, golemSymbol string, marketPrice int) () {
	changeOrderStatus(order, "InExecuting")
	// Update inventories with new contents
	// Remove from golem inventory
	newInventory := schema.Inventory{
		LocationSymbol: golemSymbol,
		Contents: make(map[string]int),
	}

	userData.Inventories[golemSymbol].Contents[order.Order.ItemSymbol] -= order.Order.Quantity
	
	log.RoutineDebug.Printf("[%s] newQuant: %d", rid, userData.Inventories[golemSymbol].Contents[order.Order.ItemSymbol])
	// delete symbol in contents if empty, then if contents empty delete entry in Inventories
	if userData.Inventories[golemSymbol].Contents[order.Order.ItemSymbol] == 0 {
		log.RoutineDebug.Printf("[%s] Deleting inventory entry for %s, as it is 0", rid, order.Order.ItemSymbol)
		delete(userData.Inventories[golemSymbol].Contents, order.Order.ItemSymbol)
		if len(userData.Inventories[golemSymbol].Contents) == 0 {
			log.RoutineDebug.Printf("[%s] Deleting golem inventory, as it is empty", rid)
			delete(userData.Inventories, golemSymbol)
			deleteGolemInventory(userData.Token, getInventoryPath(golemSymbol))
		} else {
			log.RoutineDebug.Printf("[%s] Saving updated inventory", rid)
			newInventory = userData.Inventories[golemSymbol]
			updateGolemInventory(userData.Token, getInventoryPath(golemSymbol), newInventory)
		}
	} else {
		log.RoutineDebug.Printf("[%s] Saving updated inventory", rid)
		newInventory = userData.Inventories[golemSymbol]
		updateGolemInventory(userData.Token, getInventoryPath(golemSymbol), newInventory)
	}

	// Add to market stock
	market.Stock[order.Order.ItemSymbol] += order.Order.Quantity
	// Give user coins
	userData.Coins += uint64(marketPrice * order.Order.Quantity)
	// Change golem status and statusdetail
	golem, ok := userData.Golems[golemSymbol]
	if !ok {
		log.Error.Printf("[%s] Warning! Error while getting golem after SELL market order!", rid)
	}
	golem.Status = "idle"
	golem.StatusDetail = ""
	golem.Inventory = newInventory

	// Save both to db
	// Exclusively update the specific inventory path, coin path, and golem path in the user data to avoid race conditions
	// inventoryPath := fmt.Sprintf("[\"inventories\"][\"%s\"]", golemSymbol)
	saveCoinsErr := schema.SaveUserDataAtPathToDB(UserDatabase, userData.Token, ".coins", userData.Coins)
	if saveCoinsErr != nil {
		log.Error.Printf("[%s] Warning! Error while saving after SELL market order: %v", rid, saveCoinsErr)
	}
	saveGolem(rid, userData.Token, golem)
	saveMarketErr := schema.Market_save_to_db(WorldDatabase, market)
	if saveMarketErr != nil {
		log.Error.Printf("[%s] Warning! Error while saving after SELL market order: %v", rid, saveMarketErr)
	}
	// Done
	changeOrderStatus(order, "Executed")
}