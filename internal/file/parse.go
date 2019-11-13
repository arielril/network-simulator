package file

import (
	"fmt"

	. "github.com/arielril/network-simulator/internal/network"
	fp "github.com/novalagung/gubrak"
)

const (
	NODE_LABEL         = "#NODE"
	ROUTER_LABEL       = "#ROUTER"
	ROUTER_TABLE_LABEL = "#ROUTERTABLE"
)

func findLabelIndex(lb string, list []string) (int, error) {
	return fp.FindIndex(list, func(val string) bool {
		return val == lb
	})
}

// Create nodes from parsed file
func createNodes(lines []string) {
	startIdx, _ := findLabelIndex(NODE_LABEL, lines)
	fmt.Println("Node idx:", startIdx)
}

// Create the router table for the router from parsed file
func createRouterTable(lines []string, routerName string) {
	startIdx, _ := findLabelIndex(ROUTER_LABEL, lines)
	fmt.Println("Router Table idx:", startIdx)
}

// Create routers from parsed file
func createRouters(lines []string) {
	startIdx, _ := findLabelIndex(ROUTER_LABEL, lines)
	fmt.Println("Router idx:", startIdx)
}

func Parse(lines []string) *Network {
	return &Network{}
}
