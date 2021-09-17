package main

import (
	"github.com/brct-james/brct-io-game/log"
)

// Configuration

var wipeDatabases bool = true
var refreshAuthSecret bool = true

// Global Vars

var apiVersion string = "v0.0.1"
var (
	ListenAddr = "localhost:50235"
	RedisAddr = "localhost:6381"
)

//Main
func main() {
	log.Info.Println("Brct-Game Rest API Server ", apiVersion)
}