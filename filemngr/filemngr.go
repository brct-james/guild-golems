package filemngr

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/brct-james/brct-game/log"
)

// Ensure file exists, if not create it
func Touch(name string) error {
	log.Debug.Printf("Ensuring %s exists", name)
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		// Depending on the file different responses may be valid - pass errors up the stack
		return err
	}
	return file.Close()
}

// Reads file at string path to a slice of strings by line
func ReadFileToLineSlice(filePath string) ([]string, error) {
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		// Depending on the file different responses may be valid - pass errors up the stack
		return nil, err
	}
	lines := strings.Split(string(input), "\n")
	return lines, nil
}

// Search slice for search key, returns true, index if found, else false
func KeyInSliceOfLines(searchKey string, lines []string) (bool, int) {
	for i, line := range lines {
		if strings.Contains(line, searchKey) {
			log.Debug.Printf("Found search key %s at line: %v", searchKey, i)
			return true, i
		}
	}
	return false, 0
}

// Write slice of lines to file at path
func WriteLinesToFile(filePath string, lines []string) error {
	output := strings.Join(lines, "\n")
	err := ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		// Depending on the file different responses may be valid - pass errors up the stack
		return err
	}
	return nil
}