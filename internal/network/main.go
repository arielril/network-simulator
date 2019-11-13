package network

import (
	"fmt"

	. "github.com/arielril/network-simulator/internal/component"
	"github.com/arielril/network-simulator/internal/file"
	fp "github.com/novalagung/gubrak"
)

type Network struct {
	Nodes   []Node
	Routers []Router
}

type INetwork interface {
	SendMsg(srcNode Node, dstNode Node, msg string)
}

func (net Network) SendMsg(srcNode Node, dstNode Node, msg string) {
	fmt.Printf("From: %v\n", srcNode.Name)
	fmt.Printf("To: %v\n", dstNode.Name)
	fmt.Printf("   Message: %v\n", msg)
}

func CreateNetwork(fileLines []string) Network {
	// Create nodes
	nodes := file.CreateNodes(fileLines)

	// Create routers
	routers := file.CreateRouters(fileLines)

	// Add router table for the routers
	routersAndTables, _ := fp.Map(
		routers,
		func(router Router) Router {
			rt := file.CreateRouterTable(fileLines, router.Name)
			return router.SetRouterTable(rt)
		},
	)
	routers = routersAndTables.([]Router)

	return Network{
		Nodes:   nodes,
		Routers: routers,
	}
}
