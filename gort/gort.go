// Package gort provides a way to create and manage goroutines
package gort

import "github.com/brct-james/guild-golems/log"

var Quit chan bool = make(chan bool)

func InvokeContinuous(sourceRoutine func(args ...interface{}), args ...interface{}) {
	routine := func(quit chan bool, args ...interface{}) {
		for {
			select {
			case <-quit:
				log.Info.Printf("Quit Signal Received")
				return
			default:
				sourceRoutine(args)
			}
		}
	}
	go routine(Quit, args)
}

func Invoke(routine func(args ...interface{}), args ...interface{}) {
	go routine(args)
}