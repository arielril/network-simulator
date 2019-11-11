package file

import (
	"bufio"
	"fmt"
	"os"
)

func Read(path string) []string {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		panic(
			fmt.Sprintf("Failed to open file: %v", err),
		)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
