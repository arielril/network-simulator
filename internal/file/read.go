package file

import (
	"bufio"
	"os"
)

func Read(path string) []string {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		panic("Failed to open file")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
