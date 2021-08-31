package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/brct-james/guild-golems/db"
	"github.com/gorilla/mux"
)

var apiVersion string = "v0"

var (
	ListenAddr = "localhost:50242"
	RedisAddr = "localhost:6380"
)

var dbMap = map[string]int{
	"users": 0,
	"world": 1,
}

func main() {
	fmt.Println("Guild Golems Rest API Server v0.0.1")
	fmt.Println("Connecting to Redis DB")
	// fmt.Println(dbMap["users"])
	// db.NewDatabase(RedisAddr, 0)
	udb := db.NewDatabase(RedisAddr, dbMap["users"])
	db.SetUser(udb, "testUser", "token", 0)
	db.SetUser(udb, "Greenitthe", "token", 0)
	db.GetUser(udb, "testUser")
	db.GetUser(udb, "Greenitthe")
	wdb := db.NewDatabase(RedisAddr, dbMap["world"])
	fmt.Println("Loading world json")
	saveWorldJson(readJSON("./" + apiVersion + "_regions.json"), wdb)
	db.GetWorld(wdb, "ipeiros")
	handleRequests()
}

type Test struct {
	Name string `json:"name"`
	Coins string `json:"coins"`
}

func saveWorldJson(jsonBytes []byte, database db.Database) {
	fmt.Println("Unmarshaling json")
	var res db.World
	json.Unmarshal(jsonBytes, &res)
	fmt.Println("Saving json to DB")
	db.SetWorld(database, res)
}

func readJSON(path string) []byte {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully opened " + path)
	defer jsonFile.Close()
	fmt.Println("Reading from file")
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue
}

func handleRequests() {
	//mux router
	mxr := mux.NewRouter().StrictSlash(true)
	mxr.HandleFunc("/", homepage)
	mxr.HandleFunc("/api", apiSelection)
	mxr.HandleFunc("/api/v0", v0Docs)
	mxr.HandleFunc("/api/v0/account", accountInfo)
	mxr.HandleFunc("/api/v0/locations", locationsOverview)
	mxr.HandleFunc("/api/v0/finances", financesOverview)
	mxr.HandleFunc("/api/v0/rituals", ritualsOverview)
	mxr.HandleFunc("/api/v0/guilds", guildsOverview)
	fmt.Println("Listening on :50242")
	log.Fatal(http.ListenAndServe(":50242", mxr))
}

func homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Guild Golems")
	fmt.Println("Hit: homepage")
}

func apiSelection(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ApiSel")
	fmt.Println("Hit: apisel")
}

func v0Docs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ApiDocs")
	fmt.Println("Hit: apidocs")
}

func accountInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "accountInfo")
	fmt.Println("Hit accountInfo")
}

func locationsOverview(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "locOverview")
	fmt.Println("Hit locoverview")
}

func financesOverview(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "finOve")
	fmt.Println("Hit: finOve")
}

func ritualsOverview(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ritOve")
	fmt.Println("Hit: ritOve")
}

func guildsOverview(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "guildOve")
	fmt.Println("Hit: guildOve")
} 

// func prettyJSON(input interface{}) string {
// 	res, err := json.MarshalIndent(input, "", "  ")
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(res)
// }

// var Drones []drone
// type drone struct {
// 	Id string `json:"id"`
// 	Name string `json:"name"`
// 	Model string `json:"model"`
// 	CurrentFuel uint32 `json:"currentFuel"`
// }

// func returnAllDrones(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("Endpoint Hit: returnAllDrones")
// 	fmt.Fprint(w, prettyJSON(Drones))
// }

// func returnSingleDrone(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	key := vars["id"]
// 	fmt.Println("Endpoint Hit: returnSingleDrone, key: " + key)

// 	for _, drone := range Drones {
// 		if drone.Id == key {
// 			fmt.Fprint(w, prettyJSON(drone))
// 		}
// 	}
// }

// var vs venus
// func returnVenusStatus(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("Endpoint Hit: returnVenusStatus")
// 	// [dep] Serve json without indent
// 	// json.NewEncoder(w).Encode(vs)
// 	w.Header().Set("Content-Type", "application/json")
// 	// Serve json with indent
// 	fmt.Fprint(w, prettyJSON(vs))
// }

// type venus struct {
// 	MultiverseVersion string `json:"MultiverseVersion"`
// 	VenusianId uint8 `json:"VenusianId"`
// 	Atmosphere venusianAtmosphere `json:"Atmosphere"`
// 	Surface venusianSurface `json:"Surface"`
// }

// type venusianAtmosphere struct {
// 	Pressure float64 `json:"Pressure"`
// }

// type venusianSurface struct {
// 	Temperature float64 `json:"Temperature"`
// }

// func createNewDrone(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("Endpoint Hit: createNewDrone")
// 	// get the body of the POST request
// 	// return the string response containing the request body
// 	reqBody, _ := ioutil.ReadAll(r.Body)
// 	var newDrone drone
// 	json.Unmarshal(reqBody, &newDrone)
// 	index := getStructByFieldValue(Drones, "Id", newDrone.Id)
// 	if (index != -1) {
// 		fmt.Fprintf(w, "{\"error\": \"Could not execute CREATE as drone with id %s already exists\"}", newDrone.Id)
// 	} else {
// 		Drones = append(Drones, newDrone)
// 		fmt.Fprint(w, prettyJSON(Drones))
// 	}
// }

// func deleteDrone(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id := vars["id"]
// 	fmt.Println("Endpoint Hit: deleteDrone, id: " + id)

// 	index := getStructByFieldValue(Drones, "Id", id)
// 	if (index == -1) {
// 		fmt.Fprintf(w, "{\"error\": \"Could not execute DELETE as drone with id %s does not exist\"}", id)
// 	} else {
// 		Drones = append(Drones[:index], Drones[index+1:]...)
// 		fmt.Fprint(w, prettyJSON(Drones))
// 	}
// }

// func updateDrone(w http.ResponseWriter, r *http.Request) {
// 	reqBody, _ := ioutil.ReadAll(r.Body)
// 	var newDrone drone
// 	json.Unmarshal(reqBody, &newDrone)
	
// 	vars := mux.Vars(r)
// 	id := vars["id"]
// 	fmt.Println("Endpoint Hit: updateDrone, id: " + id)

// 	index := getStructByFieldValue(Drones, "Id", id)
// 	if (index == -1) {
// 		fmt.Fprintf(w, "{\"error\": \"Could not execute UPDATE as drone with id %s does not exist\"}", id)
// 	} else {
// 		Drones[index] = newDrone
// 		fmt.Fprint(w, prettyJSON(Drones))
// 	}
// }

// func getStructByFieldValue(slice interface{} ,fieldName string,fieldValueToCheck interface {}) int {
// 	// Check for value of a given field in a slice of structs
// 	rangeOnMe := reflect.ValueOf(slice)
// 	for i := 0; i < rangeOnMe.Len(); i++ {
// 		s := rangeOnMe.Index(i)
// 		f := s.FieldByName(fieldName)
// 		if f.IsValid(){
// 			if f.Interface() == fieldValueToCheck {
// 				return i
// 			}
// 		}
// 	}
// 	return -1
// }