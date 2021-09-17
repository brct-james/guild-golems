package log

import (
	"log"
	"os"
)

// Info writes logs in the color blue with "INFO: " as prefix
var Info = log.New(os.Stdout, Blue("INFO: "), log.LstdFlags)

// Warning writes logs in the color yellow with "WARNING: " as prefix
var Warning = log.New(os.Stdout, Yellow("WARNING: "), log.LstdFlags|log.Lshortfile)

// Error writes logs in the color red with "ERROR: " as prefix
var Error = log.New(os.Stdout, Red("ERROR: "), log.LstdFlags|log.Lshortfile)

// Debug writes logs in the color cyan with "DEBUG: " as prefix
var Verbose = log.New(os.Stdout, Cyan("VERBOSE: "), log.LstdFlags|log.Lshortfile)

func Blue(in string) string {
	return "\u001b[34m" + in + "\u001b[0m"
}

func Yellow(in string) string {
	return "\u001b[33m" + in + "\u001b[0m"
}

func Red(in string) string {
	return "\u001b[31m" + in + "\u001b[0m"
}

func Cyan(in string) string {
	return "\u001b[36m" + in + "\u001b[0m"
}