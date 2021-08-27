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