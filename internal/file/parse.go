package file

import (
	"fmt"
	"strconv"
	"strings"

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

func getIp(str string) component.IP {
	values := strings.Split(str, "/")
	var prefix uint8

	if len(values) > 1 {
		pref, _ := strconv.ParseUint(values[1], 10, 8)
		prefix = uint8(pref)
	}

	return component.IP{
		Ip:     values[0],
		Prefix: prefix,
	}
}

func getNode(line string) component.Node {
	ln := strings.Split(line, ",")

	mtu, _ := strconv.ParseUint(ln[3], 10, 8)
	return component.Node{
		Name:    strings.ToUpper(ln[0]),
		Gateway: getIp(ln[4]),
		NetInt: component.NetInterface{
			Mac: ln[1],
			Ip:  getIp(ln[2]),
			Mtu: uint8(mtu),
		},
	}
}

// Create nodes from parsed file
func CreateNodes(lines []string) []component.Node {
	startIdx, _ := findLabelIndex(NODE_LABEL, lines)

	var nodeList []component.Node = make([]component.Node, 0)
	for i := startIdx + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "#") {
			break
		}

		nodeList = append(nodeList, getNode(lines[i]))
	}

	return nodeList
}

// Create the router table for the router from parsed file
func CreateRouterTable(lines []string, router string) component.RouterTable {
	startIdx, _ := findLabelIndex(ROUTER_TABLE_LABEL, lines)
	fmt.Println("Router Table idx:", startIdx)
	return component.RouterTable{
		Name: "RouterTable",
	}
}

func getRouter(line string) component.Router {
	// <router_name>,<num_ports>,<MAC0>,<IP0/prefix>,<MTU0>,<MAC1>,<IP1/prefix>,<MTU1>,<MAC2>,<IP2/prefix>,<MTU2> …
	ln := strings.Split(line, ",")

	return component.Router{
		Name: strings.ToUpper(ln[0]),
	}
}

// Create routers from parsed file
func CreateRouters(lines []string) []component.Router {
	startIdx, _ := findLabelIndex(ROUTER_LABEL, lines)
	fmt.Println("Router idx:", startIdx)

	routerList := make([]component.Router, 0)
	for i := startIdx + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "#") {
			break
		}

		routerList = append(routerList, getRouter(lines[i]))
	}

	return routerList
}
