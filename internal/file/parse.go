package file

import (
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

func createArpTb() component.ArpTable {
	tb := make(map[component.IP]string, 0)
	return component.ArpTable{
		Table: tb,
	}
}

func getNode(line string) component.Node {
	ln := strings.Split(line, ",")

	mtu, _ := strconv.ParseUint(ln[3], 10, 8)
	return component.Node{
		Name:    ln[0],
		Gateway: getIp(ln[4]),
		NetInt: component.NetInterface{
			Mac: ln[1],
			Ip:  getIp(ln[2]),
			Mtu: uint8(mtu),
		},
		ArpTable: createArpTb(),
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

func getRouterTableLines(lines []string) []string {
	startIdx, _ := findLabelIndex(ROUTER_TABLE_LABEL, lines)

	routerTableLines := make([]string, 0)
	for i := startIdx + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "#") {
			break
		}
		routerTableLines = append(routerTableLines, lines[i])
	}
	return routerTableLines
}

// Create the router table for the router from parsed file
func CreateRouterTable(lines []string, router component.Router) map[component.IP]component.RouterTableEntry {
	routerTableLines := getRouterTableLines(lines)

	mappedLines, _ := fp.Map(
		routerTableLines,
		func(line string) []string {
			return strings.Split(line, ",")
		},
	)

	goodValues, _ := fp.Filter(
		mappedLines,
		func(line []string) bool {
			return line[0] == router.Name
		},
	)

	rt := make(map[component.IP]component.RouterTableEntry, len(goodValues.([][]string)))

	for _, val := range goodValues.([][]string) {
		port, _ := strconv.ParseUint(val[3], 10, 8)
		entry := component.RouterTableEntry{
			NetDest: getIp(val[1]),
			Nexthop: getIp(val[2]),
			Port:    uint8(port),
		}
		rt[entry.NetDest] = entry
	}

	return rt
}

func getPortList(qty uint8, ports []string) []component.Port {
	var portList []component.Port = make([]component.Port, qty)

	for i := 0; i < int(qty)*3; i += 3 {
		portNumber := i / 3

		mac := ports[i]
		ippref := ports[i+1]
		mtu, _ := strconv.ParseUint(ports[i+2], 10, 8)

		port := component.Port{
			Number: uint8(portNumber),
			NetInterface: component.NetInterface{
				Ip:  getIp(ippref),
				Mac: mac,
				Mtu: uint8(mtu),
			},
		}
		portList[portNumber] = port
	}

	return portList
}

func getRouter(line string) component.Router {
	ln := strings.Split(line, ",")
	numPorts, _ := strconv.ParseUint(ln[1], 10, 8)

	return component.Router{
		Name:     ln[0],
		PortList: getPortList(uint8(numPorts), ln[1:]),
	}
}

// Create routers from parsed file
func CreateRouters(lines []string) []component.Router {
	startIdx, _ := findLabelIndex(ROUTER_LABEL, lines)

	routerList := make([]component.Router, 0)
	for i := startIdx + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "#") {
			break
		}

		routerList = append(routerList, getRouter(lines[i]))
	}

	return routerList
}
