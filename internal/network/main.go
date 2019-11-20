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
// func (net Network) SendMsg(srcNode, dstNode Node, msg string) {
// 	arp, icmp := srcNode.SendMsg(dstNode, msg)

// 	if arp != nil {
// 		// * the dst node must answer the arp request
// 	} else if icmp != nil {
// 		// * the dst node must receive the icmp request
// 	}

// 	dstIp := dstNode.NetInt.Ip

// 	// assuming is the same net
// 	// find dstNode in the srcNode arp table
// 	dstMac := srcNode.ArpTable.GetDestination(dstIp)

// 	// has the dstNode inside the arp table
// 	if dstMac != "" {
// 		// srcNode.SendIcmp(dstIp, dstMac, msg)
// 		return
// 	}

// 	// doesn't have the dstNode in the arp table
// 	// err := net.ReqArp(srcNode, dstNode)

// 	if err != nil {
// 		panic("Failed to find dstNode from srcNode")
// 	}

// 	dstMac = srcNode.ArpTable.GetDestination(dstIp)
// 	// srcNode.SendIcmp(dstMac, msg)
// 	return
// }

func (net Network) ReqArp(srcNode Node, dstNode Node) {
	// if dstMac == "" {
	// dstMac = "0xFFFFFFFF"
	// }

	// send arp request
	// fmt.Printf(
	// 	"%v box %v : ETH (src=%v dst =%v) \n ARP - Who has %v? Tell %v;",
	// 	srcNode.Name, srcNode.Name, srcNode.NetInt.Mac, dstMac, dstIp, srcNode.NetInt.Ip.ToString(),
	// )
}

func (net Network) ResArp(srcName, srcIp, srcMac, dstName, dstIp, dstMac string) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \n ARP - %v is at %v;",
		srcName, dstName, srcMac, dstMac, srcIp, srcMac,
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

func (n Network) SendMsg(src, dst Node, msg string) {
	has := src.HasDestination(dst)

	if has {
		icmpPacket := src.CreateIcmpPacket(msg, dst)
		dst.ReceiveIcmpPacket(icmpPacket)
	} else {
		arpReqPacket := src.CreateArpReq(false, dst)
		dst.ReceiveArpPacket(arpReqPacket)
		arpResPacket := dst.CreateArpReq(true, src)
		src.ReceiveArpPacket(arpResPacket)
	}
}

func (n Network) Send(src, dst Node, msg string) {

}
