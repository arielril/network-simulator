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

/*
 ! Steps:
 * 1. Check if nodes are from the same net
 * 2. If the nodes are from same net
 * 2.1 Check if the dstNode is known in the srcNode (ArpTable)
 * 2.2 If the dstNode is known
 * 2.2.1 Send message
 * 2.3 If the dstNode is unkown
 * 2.3.1 Send ARP-req in broadcast to the net
 * 2.3.2 Receive ARP-reply from dstNode
 * 2.3.3 Save the dstNode mac's to send message
 * 2.3.4 Send message
 * 3. If the nodes aren't from the same net
 * 3.1 ...
*/
// Sends a message from srcNode to dstNode
func (net Network) SendMsg(srcNode, dstNode Node, msg string) {
	dstIp := dstNode.NetInt.Ip

	// assuming is the same net
	// find dstNode in the srcNode arp table
	dstMac := srcNode.ArpTable.GetDestination(dstIp) // this returns the dst mac address

	// has the dstNode inside the arp table
	if dstMac != "" {
		srcNode.SendIcmp(dstIp, dstMac, msg)
		return
	}

	// doesn't have the dstNode in the arp table
	err := net.SendArp(srcNode, dstIp)

	if err != nil {
		panic("Failed to find dstNode from srcNode")
	}

	dstMac = srcNode.ArpTable.GetDestination(dstIp)
	srcNode.SendIcmp(dstMac, msg)
	return
}

func (net Network) SendArp(src Node, dstIp IP) {
	// send arp request
	fmt.Printf(
		"%v box %v : ETH (src=%v dst =0xFFFFFFFF) \n ARP - Who has %v? Tell %v;",
		src.Name, src.Name, src.NetInt.Mac, dstIp.ToString(), src.NetInt.Ip.ToString(),
	)
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
			rt := file.CreateRouterTable(fileLines, router)
			return router.SetRouterTable(rt)
		},
	)
	routers = routersAndTables.([]Router)

	return Network{
		Nodes:   nodes,
		Routers: routers,
	}
}
