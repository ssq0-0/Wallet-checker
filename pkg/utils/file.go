package utils

import (
	"bufio"
	"os"
	"path/filepath"
)

// FileReader reads a file line by line and returns its contents as a slice of strings.
// It handles different line endings (\n, \r\n) and is cross-platform compatible.
func FileReader(filename string) ([]string, error) {
	filename = filepath.Clean(filename)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, err
	}

	if _, err := os.OpenFile(filename, os.O_RDONLY, 0); err != nil {
		return nil, err
	}

	var lines []string
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
