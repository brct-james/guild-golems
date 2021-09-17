package log

import (
	"io"
	"log"
	"os"
)

var (
	Info *log.Logger
	Important *log.Logger
	Test *log.Logger
	Error *log.Logger
	Debug *log.Logger
)

func init() {
	// Handle logging to file
	var logpath = "./debug.ansi"
	var debugFile, logErr = os.Create(logpath)

	if logErr != nil {
		log.Fatalf("%v", logErr)
	}

	// Debug writes logs in the color cyan with "DEBUG: " as prefix
	Debug = log.New(debugFile, Cyan("DEBUG: "), log.LstdFlags|log.Lshortfile)

	multiOut := io.MultiWriter(os.Stdout, debugFile)

	// Info writes logs in the color blue with "INFO: " as prefix
	Info = log.New(multiOut, Blue("INFO: "), log.LstdFlags)

	// Important writes logs in the color yellow with "IMPORTANT: " as prefix
	Important = log.New(multiOut, Yellow("IMPORTANT: "), log.LstdFlags|log.Lshortfile)

	// Test writes logs in the color White on Magenta Background with "TEST: " as prefix
	Test = log.New(multiOut, White(CyanBackground("TEST:")) + " ", log.LstdFlags|log.Lshortfile)

	// Error writes logs in the color Red with "ERROR: " as prefix
	Error = log.New(multiOut, Red("ERROR: "), log.LstdFlags|log.Lshortfile)
}

// Formatting functions

func Bold(in string) string {
	return "\u001b[1m" + in + "\u001b[0m"
}

//Coloring functions

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

func White(in string) string {
	return "\u001b[37m" + in + "\u001b[0m"
}

func Green(in string) string {
	return "\u001b[32m" + in + "\u001b[0m"
}

func MagentaBackground(in string) string {
	return "\u001b[45m" + in + "\u001b[0m"
}

func CyanBackground(in string) string {
	return "\u001b[46m" + in + "\u001b[0m"
}

func TestSuccess(in string) string {
	return "\u001b[42m\u001b[39m" + in + "\u001b[0m"
}

func TestFail(in string) string {
	return "\u001b[41m\u001b[39m" + in + "\u001b[0m"
}

func TestOutput(in string, successString string) string {
	if in == successString {
		return TestSuccess(in)
	} else {
		return TestFail(in)
	}
}