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

type packetType uint8

const (
	ARP_REQ packetType = iota + 1
	ARP_REP
	ICMP_REQ
	ICMP_REP
)

var lastIcmpPacket packet
var globalSrcNode NetComponent

type MAC string
type MTU uint8
type IP struct {
	ip     string
	prefix uint8
}

func NewIp(ip string) *IP {
	ipSplit := strings.Split(ip, "/")
	var prefix uint8

	if len(ipSplit) == 1 {
		prefix = 0
	} else {
		pref, _ := strconv.ParseUint(ipSplit[1], 10, 8)
		prefix = uint8(pref)
	}

	netIp := &IP{
		ip:     ipSplit[0],
		prefix: prefix,
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
	typ  packetType
}

func NewPacket(src, dst packetHost, typ packetType, data string, ttl, mf, off uint8) packet {
	pkt := packet{
		src:  src,
		dst:  dst,
		typ:  typ,
		data: data,
		ttl:  ttl,
		mf:   mf,
		off:  off,
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
	return NewPacket(srcHost, dstHost, ARP_REP, "", 0, 0, 0)
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
	return NewPacket(srcHost, dstHost, ICMP_REQ, data, 8, 0, 0)
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

type NetComponent interface {
	GetName() string
	SendMessage(msg string, dest NetComponent, destNetInterface netInterface, env Environment)
	SendArpReply(pkt packet) packet
	ReceiveArpRequest(pkt packet)
	SendIcmpReply(pkt packet) packet
	ReceiveIcmpRequest(pkt packet, env Environment)
	ReceiveIcmpReply(pkt packet, env Environment)
}

type Node interface {
	GetNetInterface() netInterface
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
	arpTb := make(map[IP]MAC)
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

func (n *node) GetName() string {
	return n.name
}

func (n *node) GetNetInterface() netInterface {
	return n.netPort
}

func (n *node) ReceiveArpRequest(pkt packet) {
	_, hasSrc := n.arpTable[pkt.src.ip]

	if !hasSrc {
		n.arpTable[pkt.src.ip] = pkt.src.mac
	}
}

func (n *node) SendArpReply(pkt packet) packet {
	srcHost := packetHost{
		name: n.name,
		ip:   n.netPort.ip,
		mac:  n.netPort.mac,
	}
	return NewPacket(srcHost, pkt.src, ARP_REP, "", 8, 0, 0)
}

func (n *node) ReceiveIcmpRequest(pkt packet, env Environment) {
	fmt.Printf("%v rbox %v : Received %v;\n", n.name, n.name, pkt.data)
}

func (n *node) ReceiveIcmpReply(pkt packet, env Environment) {
	fmt.Printf("%v rbox %v : Received %v;\n", n.name, n.name, pkt.data)
}

func (n *node) SendIcmpReply(pkt packet) packet {
	return NewPacket(pkt.dst, pkt.src, ICMP_REP, pkt.data, 8, 0, 0)
}

func (n *node) SendMessage(msg string, dest NetComponent, destNetInterface netInterface, env Environment) {
	isSameNet := n.netPort.ip.IsSameNet(destNetInterface.ip)

	var dstNetPort netInterface
	var dstName string
	var arpTbIpSearch IP

	if !isSameNet {
		rt := env.GetDefaultGateway(n)
		prt, _ := fp.Find(rt.ports, func(p routerPort) bool {
			return p.ip == n.gateway
		})

		dstNetPort = netInterface{
			ip:  destNetInterface.ip,
			mac: (prt.(routerPort)).mac,
			mtu: (prt.(routerPort)).mtu,
		}

		dstName = rt.name
		routerPortIp := (prt.(routerPort)).ip
		arpTbIpSearch = routerPortIp
	} else {
		dstNetPort = destNetInterface
		arpTbIpSearch = dstNetPort.ip
		dstName = dest.GetName()
	}

	_, hasMac := n.arpTable[arpTbIpSearch]

	if !hasMac {
		pkt := createBroadcastArpReq(n.name, n.netPort, arpTbIpSearch)
		arpReply := env.SendArpReq(pkt)
		n.ReceiveArpRequest(arpReply)
	}

	pkt := createIcmpReq(n.name, dstName, n.netPort, dstNetPort, msg)
	lastIcmpPacket = pkt
	icmpReply := env.SendIcmpReq(pkt)
	if icmpReply != nil {
		lastIcmpPacket = *icmpReply
		n.ReceiveIcmpReply(*icmpReply, env)
	}
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
	// Arp Table
	arpTable map[IP]MAC
}

func NewRouter(name string) *router {
	ports := make([]routerPort, 0)
	routerTb := make([]*routerTableEntry, 0)
	arpTb := make(map[IP]MAC)
	return &router{
		name:        name,
		ports:       ports,
		routerTable: routerTb,
		arpTable:    arpTb,
	}
}

func (r *router) GetName() string {
	return r.name
}

func (r *router) AddPort(port routerPort) {
	r.ports = append(r.ports, port)
}

func (r *router) AddRouterTableEntry(entry *routerTableEntry) {
	r.routerTable = append(r.routerTable, entry)
}

func (r *router) ReceiveArpRequest(pkt packet) {
	_, hasSrcIp := r.arpTable[pkt.src.ip]

	if !hasSrcIp {
		r.arpTable[pkt.src.ip] = pkt.src.mac
	}
}

func (r *router) GetPortByMac(mac MAC) routerPort {
	rp, _ := fp.Find(r.ports, func(prt routerPort) bool {
		return prt.mac == mac
	})
	return rp.(routerPort)
}

func (r *router) GetPortByIp(ip IP) routerPort {
	rp, _ := fp.Find(r.ports, func(prt routerPort) bool {
		return prt.ip == ip
	})
	return rp.(routerPort)
}

func (r *router) SendArpReply(pkt packet) packet {
	dstIp := pkt.dst.ip
	dstPort := r.GetPortByIp(dstIp)
	srcHost := packetHost{
		name: r.name,
		ip:   dstPort.ip,
		mac:  dstPort.mac,
	}
	return NewPacket(srcHost, pkt.src, ARP_REP, "", 8, 0, 0)
}

func (r *router) ReceiveIcmpRequest(pkt packet, env Environment) {
	port := r.GetPortByMac(pkt.dst.mac)
	_ = env.SendMessage(pkt.data, port.ip, pkt.dst.ip)
}

func (r *router) ReceiveIcmpReply(pkt packet, env Environment) {
	port := r.GetPortByMac(pkt.dst.mac)
	_ = env.SendMessage(pkt.data, port.ip, pkt.dst.ip)
}

func (r *router) SendIcmpReply(pkt packet) packet {
	port := r.GetPortByMac(pkt.dst.mac)
	srcHost := packetHost{
		name: r.name,
		ip:   port.ip,
		mac:  port.mac,
	}
	return NewPacket(srcHost, pkt.src, ICMP_REP, pkt.data, 8, 0, 0)
}

func (r *router) SendMessage(msg string, dest NetComponent, destNetInterface netInterface, env Environment) {
	// verify if the router can reach the network
	entry, _ := fp.Find(r.routerTable, func(e *routerTableEntry) bool {
		return e.netdest.IsSameNet(destNetInterface.ip)
	})

	if entry == nil {
		panic("Invalid network. No entry in the router table")
	}

	// retrieve the port that can reach the netork
	port, _ := fp.Find(r.ports, func(p routerPort) bool {
		return p.number == entry.(*routerTableEntry).port
	})
	networkPort := port.(routerPort).netInterface

	// verify if the destination is known by the router
	_, hasMacArpTable := r.arpTable[destNetInterface.ip]
	if !hasMacArpTable {
		pkt := createBroadcastArpReq(r.name, networkPort, destNetInterface.ip)
		arpReply := env.SendArpReq(pkt)
		r.ReceiveArpRequest(arpReply)
	}

	pkt := createIcmpReq(r.name, dest.GetName(), networkPort, destNetInterface, msg)
	pkt.src.ip = lastIcmpPacket.src.ip
	pkt.ttl--
	lastIcmpPacket = pkt
	icmpReply := env.SendIcmpReq(pkt)
	if icmpReply != nil {
		lastIcmpPacket = *icmpReply
		r.ReceiveIcmpRequest(*icmpReply, env)
	}
}

type Environment interface {
	AddNode(nd *node)
	AddRouter(r *router)
	GetDefaultGateway(n *node) *router
	GetNetComponentByName(name string) NetComponent
	ParseLines(lines []string)

	SendMessage(msg string, ipSrc, ipDest IP) error
	SendArpReq(pkt packet) packet
	SendIcmpReq(pkt packet) *packet
}

type environment struct {
	nodes   []*node
	routers []*router
}

func NewEnvironment() Environment {
	nodeList := make([]*node, 0)
	routerList := make([]*router, 0)
	return &environment{
		nodes:   nodeList,
		routers: routerList,
	}
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

func (e *environment) GetRouterByName(name string) *router {
	for _, r := range e.routers {
		if r.name == name {
			return r
		}
	}
	return nil
}

func (e *environment) GetNetComponentByMac(mac MAC) NetComponent {
	var component NetComponent

	for _, n := range e.nodes {
		if n.netPort.mac == mac {
			component = n
			break
		}
	}

	for _, r := range e.routers {
		for _, rp := range r.ports {
			if rp.mac == mac {
				component = r
				break
			}
		}
	}

	return component
}

func (e *environment) GetNetComponentByIp(ip IP) NetComponent {
	var component NetComponent

	for _, n := range e.nodes {
		if n.netPort.ip == ip {
			component = n
			break
		}
	}

	for _, r := range e.routers {
		for _, rp := range r.ports {
			if rp.ip == ip {
				component = r
				break
			}
		}
	}

	return component
}

func (e *environment) GetNetComponentByName(name string) NetComponent {
	var comp NetComponent

	for _, n := range e.nodes {
		if n.name == name {
			comp = n
		}
	}

	for _, r := range e.routers {
		if r.name == name {
			comp = r
		}
	}

	return comp
}

func (e *environment) SendArpReq(pkt packet) packet {
	logArpReq(pkt)
	dst := e.GetNetComponentByIp(pkt.dst.ip)
	dst.ReceiveArpRequest(pkt)
	arpReply := dst.SendArpReply(pkt)
	logArpRep(arpReply)
	return arpReply
}

func (e *environment) SendIcmpReq(pkt packet) *packet {
	logIcmpReq(pkt)
	dst := e.GetNetComponentByMac(pkt.dst.mac)
	dst.ReceiveIcmpRequest(pkt, e)
	// the origin component doesnt set the icmp reply
	if dst.GetName() != globalSrcNode.GetName() {
		icmpRep := dst.SendIcmpReply(pkt)
		logIcmpRep(icmpRep)
		return &icmpRep
	}
	return nil
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

func (e *environment) GetComponentNetInterfaceByIp(comp NetComponent, ip IP) netInterface {
	switch comp := comp.(type) {
	case *node:
		return comp.GetNetInterface()
	case *router:
		for _, port := range comp.ports {
			if port.ip == ip {
				return port.netInterface
			}
		}
	}
	return netInterface{}
}

func (e *environment) SendMessage(msg string, ipSrc, ipDest IP) error {
	src := e.GetNetComponentByIp(ipSrc)
	dst := e.GetNetComponentByIp(ipDest)
	if src == nil || dst == nil {
		return fmt.Errorf(
			"Invalid source (%v) or destination (%v) to send message",
			src, dst,
		)
	}

	destNetInterface := e.GetComponentNetInterfaceByIp(dst, ipDest)
	src.SendMessage(msg, dst, destNetInterface, e)
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

func parseRouterTableEntry(line string) (string, *routerTableEntry) {
	l := strings.Split(line, ",")
	routerName := l[0]

	netDest := *NewIp(l[1])
	nexthop := *NewIp(l[2])
	port, _ := strconv.ParseUint(l[3], 10, 8)

	return routerName, &routerTableEntry{
		netdest: netDest,
		nexthop: nexthop,
		port:    uint8(port),
	}
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

	routeTableIdx, _ := findLabelIndex(ROUTER_TABLE_LABEL, lines)
	for i := routeTableIdx + 1; i < lenLines; i++ {
		if strings.Contains(lines[i], "#") {
			break
		}
		routerName, entry := parseRouterTableEntry(lines[i])
		router := e.GetRouterByName(routerName)
		router.AddRouterTableEntry(entry)
	}
}

func Run(ctx *cli.Context) error {
	args := &file.InputArgs{}

	_ = file.ValidateInputeArgs(args, ctx)

	filePath, _ := filepath.Abs(args.Topology)
	fileR := file.Read(filePath)

	// craete env and parse lines
	env := NewEnvironment()
	env.ParseLines(fileR)

	src := env.GetNetComponentByName(args.SrcNode)
	ipSrc := src.(Node).GetNetInterface()
	globalSrcNode = src

	dest := env.GetNetComponentByName(args.DstNode)
	ipDest := dest.(Node).GetNetInterface()

	return env.SendMessage(args.Msg, ipSrc.ip, ipDest.ip)
}
