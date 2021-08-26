package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Guild Golems Rest API Server v0.0.1")
	handleRequests()
}

func handleRequests() {
	//mux router
	mxr := mux.NewRouter().StrictSlash(true)
	mxr.HandleFunc("/", homepage)
	fmt.Println("Listening on :50242")
	log.Fatal(http.ListenAndServe(":50242", mxr))
}

func homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Guild Golems")
	fmt.Println("Endpoint Hit: homepage")
}

// func prettyJSON(input interface{}) string {
// 	res, err := json.MarshalIndent(input, "", "  ")
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(res)
// }