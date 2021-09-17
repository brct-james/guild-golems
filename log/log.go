package log

import (
	"io"
	"log"
	"os"
)

var (
	Info *log.Logger
	Important *log.Logger
	Error *log.Logger
	Debug *log.Logger
)

func init() {
	// Handle logging to file
	var logpath = "./debug.log"
	var debugFile, logErr = os.Create(logpath)

	if logErr != nil {
		panic(logErr)
	}
	defer debugFile.Close()

	// Debug writes logs in the color cyan with "DEBUG: " as prefix
	Debug = log.New(debugFile, Cyan("DEBUG: "), log.LstdFlags|log.Lshortfile)

	multiOut := io.MultiWriter(os.Stdout, debugFile)

	// Info writes logs in the color blue with "INFO: " as prefix
	Info = log.New(multiOut, Blue("INFO: "), log.LstdFlags)

	// Important writes logs in the color yellow with "Important: " as prefix
	Important = log.New(multiOut, Yellow("IMPORTANT: "), log.LstdFlags|log.Lshortfile)

	// Error writes logs in the color red with "ERROR: " as prefix
	Error = log.New(multiOut, Red("ERROR: "), log.LstdFlags|log.Lshortfile)
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