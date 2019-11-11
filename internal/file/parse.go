package file

import (
	"fmt"
)

func Parse(lines []string) {
	for i, line := range lines {
		fmt.Printf("Line %d: %v\n", i+1, line)
	}
}
