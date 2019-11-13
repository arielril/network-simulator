package file

import (
	"fmt"

	"github.com/arielril/network-simulator/internal/component"

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
func CreateNodes(lines []string) []component.Node {
	startIdx, _ := findLabelIndex(NODE_LABEL, lines)
	fmt.Println("Node idx:", startIdx)

	l := make([]component.Node, 1)
	l[0] = component.Node{
		Name: "Node",
	}
	return l
}

// Create the router table for the router from parsed file
func CreateRouterTable(lines []string, router string) component.RouterTable {
	startIdx, _ := findLabelIndex(ROUTER_TABLE_LABEL, lines)
	fmt.Println("Router Table idx:", startIdx)
	return component.RouterTable{
		Name: "RouterTable",
	}
}

// Create routers from parsed file
func CreateRouters(lines []string) []component.Router {
	startIdx, _ := findLabelIndex(ROUTER_LABEL, lines)
	fmt.Println("Router idx:", startIdx)

	l := make([]component.Router, 1)
	l[0] = component.Router{
		Name: "Router",
	}
	return l
}
