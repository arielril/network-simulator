package simulator

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arielril/network-simulator/internal/file"
	"github.com/urfave/cli"

	fp "github.com/novalagung/gubrak"
)

const (
	UNKOWN_MAC         MAC    = "FF:FF:FF:FF:FF:FF"
	NODE_LABEL         string = "#NODE"
	ROUTER_LABEL       string = "#ROUTER"
	ROUTER_TABLE_LABEL string = "#ROUTERTABLE"
	MASK               uint32 = 0xFFFFFFFF
)

type MAC string
type MTU uint8
type IP struct {
	ip     string
	prefix uint8
}

func NewIp(ip string) *IP {
	ipSplit := strings.Split(ip, "/")
	ipPref, _ := strconv.ParseUint(ipSplit[1], 10, 8)
	netIp := &IP{
		ip:     ipSplit[0],
		prefix: uint8(ipPref),
	}
	return netIp
}

func (ip IP) ToBit() uint32 {
	list, err := fp.Map(
		strings.Split(ip.ip, "."),
		func(part string) uint64 {
			val, _ := strconv.ParseUint(part, 10, 32)
			return val
		},
	)

	if err != nil {
		panic("Failed to convert IP")
	}
	ipList := list.([]uint64)
	bits := fmt.Sprintf(
		"%08b%08b%08b%08b",
		ipList[0],
		ipList[1],
		ipList[2],
		ipList[3],
	)
	res, _ := strconv.ParseUint(bits, 2, 32)
	return uint32(res)
}

func (ip IP) IsSameNet(ipDest IP) bool {
	ipSrcBit := ip.ToBit()
	ipDestBit := ipDest.ToBit()
	fst := MASK & (ipSrcBit & (MASK << (32 - ip.prefix)))
	snd := MASK & (fst | (MASK >> ip.prefix))

	return ipDestBit >= fst && ipDestBit <= snd
}

func (ip IP) ToString() string {
	return fmt.Sprintf(
		"%v/%v",
		ip.ip, ip.prefix,
	)
}

type netInterface struct {
	ip  IP
	mac MAC
	mtu MTU
}

type packetHost struct {
	name string
	mac  MAC
	ip   IP
}

type packet struct {
	src  packetHost
	dst  packetHost
	data string
	ttl  uint8
	mf   uint8
	off  uint8
}

func NewPacket(src, dst packetHost) packet {
	pkt := packet{
		src: src,
		dst: dst,
	}
	return pkt
}

func createBroadcastArpReq(srcName string, srcNetPort netInterface, ipDst IP) packet {
	srcHost := packetHost{
		name: srcName,
		ip:   srcNetPort.ip,
		mac:  srcNetPort.mac,
	}
	dstHost := packetHost{
		ip:  ipDst,
		mac: UNKOWN_MAC,
	}
	pkt := packet{
		src: srcHost,
		dst: dstHost,
	}
	return pkt
}

func createIcmpReq(srcName, dstName string, srcNetPort, dstNetPort netInterface, data string) packet {
	srcHost := packetHost{
		mac:  srcNetPort.mac,
		ip:   srcNetPort.ip,
		name: srcName,
	}
	dstHost := packetHost{
		name: dstName,
		ip:   dstNetPort.ip,
		mac:  dstNetPort.mac,
	}

	pkt := packet{
		src:  srcHost,
		dst:  dstHost,
		data: data,
		ttl:  8,
		mf:   0,
		off:  0,
	}
	return pkt
}

func logArpReq(pkt packet) {
	fmt.Printf(
		"%v box %v : ETH (src=%v dst=%v) \\n ARP - Who has %v? Tell %v;\n",
		pkt.src.name, pkt.src.name, pkt.src.mac, pkt.dst.mac,
		pkt.dst.ip.ip, pkt.src.ip.ip,
	)
}

func logArpRep(pkt packet) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \\n ARP - %v is at %v;\n",
		pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
		pkt.src.ip.ip, pkt.src.mac,
	)
}

func logIcmpReq(pkt packet) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Echo request (data=%v);\n",
		pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
		pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off, pkt.data,
	)
}

func logIcmpRep(pkt packet) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Echo reply (data=%v);\n",
		pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
		pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off, pkt.data,
	)
}

type node struct {
	// Name of the node
	name string
	// Network interface
	netPort netInterface
	// Default Gateway of the node
	gateway IP
	// Arp Table
	arpTable map[IP]MAC
}

func NewNode(name, ip, gateway string, mac MAC, mtu MTU) *node {
	ipSplit := strings.Split(ip, "/")
	ipPref, _ := strconv.ParseUint(ipSplit[1], 10, 8)
	netIp := NewIp(ip)
	netInt := netInterface{
		ip:  *netIp,
		mac: mac,
		mtu: mtu,
	}
	arpTb := make(map[IP]MAC, 0)
	nd := &node{
		name:    name,
		netPort: netInt,
		gateway: IP{
			ip:     gateway,
			prefix: uint8(ipPref),
		},
		arpTable: arpTb,
	}

	return nd
}

// Sends the msg to the dst node
func (n *node) SendMsg(msg string, dst *node, env environment) error {
	isSameNet := n.netPort.ip.IsSameNet(dst.netPort.ip)
	var dstNetPort netInterface

	if !isSameNet {
		rt := env.GetDefaultGateway(n)
		prt, _ := fp.Find(rt.ports, func(p routerPort) bool {
			return p.ip == n.gateway
		})
		dstNetPort = netInterface{
			ip:  dst.netPort.ip,
			mac: (prt.(routerPort)).mac,
			mtu: (prt.(routerPort)).mtu,
		}
		_, hasMac := n.arpTable[dstNetPort.ip]

		if !hasMac {
			pkt := createBroadcastArpReq(n.name, n.netPort, dstNetPort.ip)
			arpReply := env.SendArpReq(pkt)
			n.ReceiveArpRequest(arpReply)
		}

		pkt := createIcmpReq(n.name, dst.name, n.netPort, dstNetPort, msg)
		icmpReply := env.SendIcmpReq(pkt)
		n.ReceiveIcmpReq(icmpReply)
		return nil
	} else {
		dstNetPort = dst.netPort
	}

	_, hasMac := n.arpTable[dstNetPort.ip]

	if !hasMac {
		pkt := createBroadcastArpReq(n.name, n.netPort, dstNetPort.ip)
		arpReply := env.SendArpReq(pkt)
		n.ReceiveArpRequest(arpReply)
	}

	pkt := createIcmpReq(n.name, dst.name, n.netPort, dstNetPort, msg)
	icmpReply := env.SendIcmpReq(pkt)
	n.ReceiveIcmpReq(icmpReply)
	return nil
}

func (n *node) ReceiveArpRequest(pkt packet) {
	_, hasSrc := n.arpTable[pkt.src.ip]

	if !hasSrc {
		n.arpTable[pkt.src.ip] = pkt.src.mac
	}
}

func (n *node) ReplyArpRequest(pkt packet) packet {
	srcHost := packetHost{
		name: n.name,
		ip:   n.netPort.ip,
		mac:  n.netPort.mac,
	}
	reply := packet{
		src: srcHost,
		dst: pkt.src,
	}

	return reply
}

func (n *node) ReceiveIcmpReq(pkt packet) {
	fmt.Printf("%v rbox %v : Received %v;\n", n.name, n.name, pkt.data)
}

func (n *node) ReplyIcmpRequest(pkt packet) packet {
	reply := packet{
		src:  pkt.dst,
		dst:  pkt.src,
		data: pkt.data,
		ttl:  8,
		mf:   0,
		off:  0,
	}
	return reply
}

type routerTableEntry struct {
	netdest IP
	nexthop IP
	port    uint8
}

type routerPort struct {
	number uint8
	netInterface
}

func NewRouterPort(number uint8, ip string, mac MAC, mtu MTU) *routerPort {
	netIp := NewIp(ip)
	netInt := netInterface{
		mac: mac,
		mtu: mtu,
		ip:  *netIp,
	}
	port := &routerPort{
		number,
		netInt,
	}
	return port
}

type router struct {
	// Name of the router
	name string
	// List of ports of the router
	ports []routerPort
	// Router Table
	routerTable []*routerTableEntry
}

func NewRouter(name string) *router {
	ports := make([]routerPort, 0)
	routerTb := make([]*routerTableEntry, 0)
	rt := &router{
		name:        name,
		ports:       ports,
		routerTable: routerTb,
	}
	return rt
}

func (r *router) AddPort(port routerPort) {
	r.ports = append(r.ports, port)
}

type environment struct {
	nodes   []*node
	routers []*router
}

func NewEnvironment() environment {
	nodeList := make([]*node, 0)
	env := environment{
		nodes: nodeList,
	}
	return env
}

func (e *environment) AddNode(nd *node) {
	e.nodes = append(e.nodes, nd)
}

func (e *environment) AddRouter(rt *router) {
	e.routers = append(e.routers, rt)
}

func (e *environment) GetNodeByName(name string) *node {
	nd, _ := fp.Find(e.nodes, func(n *node) bool {
		return n.name == name
	})

	return nd.(*node)
}

func (e *environment) GetNodeByIp(ip IP) *node {
	nd, _ := fp.Find(e.nodes, func(n *node) bool {
		return n.netPort.ip.ip == ip.ip
	})
	return nd.(*node)
}

func (e *environment) GetNodeByMac(mac MAC) *node {
	nd, _ := fp.Find(e.nodes, func(n *node) bool {
		return n.netPort.mac == mac
	})
	return nd.(*node)
}

func (e *environment) SendArpReq(pkt packet) packet {
	logArpReq(pkt)
	// if pkt.dst.mac == UNKOWN_MAC {
	// }
	dstNode := e.GetNodeByIp(pkt.dst.ip)
	dstNode.ReceiveArpRequest(pkt)
	arpReply := dstNode.ReplyArpRequest(pkt)
	logArpRep(arpReply)
	return arpReply
}

func (e *environment) SendIcmpReq(pkt packet) packet {
	logIcmpReq(pkt)
	dst := e.GetNodeByMac(pkt.dst.mac)
	dst.ReceiveIcmpReq(pkt)
	icmpRep := dst.ReplyIcmpRequest(pkt)
	logIcmpRep(icmpRep)
	return icmpRep
}

func (e *environment) GetDefaultGateway(n *node) *router {
	for _, rt := range e.routers {
		for _, p := range rt.ports {
			if p.ip == n.gateway {
				return rt
			}
		}
	}
	return nil
}

func findLabelIndex(lb string, list []string) (int, error) {
	return fp.FindIndex(list, func(val string) bool {
		return val == lb
	})
}

func parseNode(line string) *node {
	l := strings.Split(line, ",")

	mtu, _ := strconv.ParseUint(l[3], 10, 8)
	mac := MAC(l[1])
	return NewNode(l[0], l[2], l[4], mac, MTU(mtu))
}

func parseRouter(line string) *router {
	l := strings.Split(line, ",")
	numPorts, _ := strconv.ParseUint(l[1], 10, 8)

	rt := NewRouter(l[0])

	portLine := l[2:]
	for i := 0; i < int(numPorts)*3; i += 3 {
		mtu, _ := strconv.ParseUint(portLine[i+2], 10, 8)
		rt.AddPort(
			*NewRouterPort(
				uint8(i/3), portLine[i+1], MAC(portLine[i]), MTU(mtu),
			),
		)
	}

	return rt
}

func (e *environment) ParseLines(lines []string) {
	nodeIdx, _ := findLabelIndex(NODE_LABEL, lines)
	lenLines := len(lines)
	for i := nodeIdx + 1; i < lenLines; i++ {
		if strings.Contains(lines[i], "#") {
			break
		}
		e.AddNode(parseNode(lines[i]))
	}

	routerIdx, _ := findLabelIndex(ROUTER_LABEL, lines)
	for i := routerIdx + 1; i < lenLines; i++ {
		if strings.Contains(lines[i], "#") {
			break
		}
		e.AddRouter(parseRouter(lines[i]))
	}
}

func Run(ctx *cli.Context) error {
	args := &file.InputArgs{}

	file.ValidateInputeArgs(args, ctx)

	filePath, _ := filepath.Abs(args.Topology)
	fileR := file.Read(filePath)

	// craete env and add nodes
	env := NewEnvironment()
	env.ParseLines(fileR)

	src := env.GetNodeByName(args.SrcNode)
	dst := env.GetNodeByName(args.DstNode)

	return src.SendMsg(args.Msg, dst, env)
}
