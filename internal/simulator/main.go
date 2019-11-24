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

type MAC string
type MTU uint8

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
	ICMP_TIME_EXCEEDED
)

type netInterface struct {
	ip  IP
	mac MAC
	mtu MTU
}

/*
----------------------------------------------------
Packet implementation
----------------------------------------------------
*/

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

func SetHosts(pkts []*packet, src, dest *packetHost) {
	for _, pkt := range pkts {
		if src != nil {
			pkt.src = *src
		}
		if dest != nil {
			pkt.dst = *dest
		}
	}
}

func Fragment(p *packet, mtu MTU) []*packet {
	dataChunk, _ := fp.Chunk(
		strings.Split(p.data, ""),
		int(mtu),
	)

	fragments, _ := fp.Map(dataChunk, func(chunk []string, i int) *packet {
		var off uint8 = p.off + (uint8(mtu) * uint8(i))
		var mf uint8

		pk := NewPacket(p.src, p.dst, p.typ, strings.Join(chunk, ""), p.ttl, mf, off)
		return &pk
	})

	frags := fragments.([]*packet)
	return frags
}

type HashMap map[string]interface{}

func reducePacket(acc HashMap, pkt *packet, i int) HashMap {
	acc["src"] = pkt.src
	acc["dst"] = pkt.dst
	acc["type"] = pkt.typ
	acc["ttl"] = pkt.ttl
	return acc
}

func Defragment(pkts []*packet) *packet {
	data := DefragmentData(pkts)

	hash, _ := fp.Reduce(pkts, reducePacket, HashMap{})
	hashMap := hash.(HashMap)

	pkt := NewPacket(
		hashMap["src"].(packetHost),
		hashMap["dst"].(packetHost),
		hashMap["type"].(packetType),
		data,
		hashMap["ttl"].(uint8),
		0, 0,
	)
	return &pkt
}

func DefragmentData(pkts []*packet) string {
	dataList, _ := fp.Map(pkts, func(p *packet) string {
		return p.data
	})
	return strings.Join(dataList.([]string), "")
}

func IsTimeExceeded(pkts []*packet) bool {
	for _, p := range pkts {
		if p.ttl == 0 {
			return true
		}
	}
	return false
}

func GetPktsDest(pkts []*packet) *packetHost {
	if len(pkts) == 0 {
		return nil
	}
	return &pkts[0].dst
}

func GetPktsSrc(pkts []*packet) *packetHost {
	if len(pkts) == 0 {
		return nil
	}
	return &pkts[0].src
}

func GetPktsType(pkts []*packet) packetType {
	if len(pkts) > 0 {
		return pkts[0].typ
	}
	return 0
}

func GetPktsTTL(pkts []*packet) uint8 {
	if len(pkts) > 0 {
		return pkts[0].ttl
	}
	return 0
}

func DecrementPktsTTL(pkts []*packet) {
	fp.ForEach(pkts, func(pkt *packet) {
		pkt.ttl--
	})
}

/*
----------------------------------------------------
Net component
----------------------------------------------------
*/

type NetComponent interface {
	GetName() string

	SendArpReply(pkt packet) packet
	SendIcmpReply(pkts []*packet, mtu MTU, env Environment) []*packet

	ReceiveArpRequest(pkt packet)
	ReceiveIcmpRequest(pkts []*packet, env Environment) bool
	ReceiveIcmpReply(pkts []*packet, env Environment)
	ReceiveTimeExceeded(pkts []*packet, env Environment)
}

/*
----------------------------------------------------
Node implementations
----------------------------------------------------
*/

type Node interface {
	GetNetInterface() netInterface
	SendMessage(msg string, dest NetComponent, destNetInterface netInterface, env Environment)
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

func (n *node) ReceiveIcmpRequest(pkt []*packet, env Environment) bool {
	data := DefragmentData(pkt)
	fmt.Printf("%v rbox %v : Received %v;\n", n.name, n.name, data)
	return true
}

func (n *node) ReceiveIcmpReply(pkt []*packet, env Environment) {
	data := DefragmentData(pkt)
	fmt.Printf("%v rbox %v : Received %v;\n", n.name, n.name, data)
}

func (n *node) ReceiveTimeExceeded(pkt []*packet, env Environment) {}

func (n *node) SendArpReply(pkt packet) packet {
	srcHost := packetHost{
		name: n.name,
		ip:   n.netPort.ip,
		mac:  n.netPort.mac,
	}
	return NewPacket(srcHost, pkt.src, ARP_REP, "", 8, 0, 0)
}

func (n *node) SendIcmpReply(pkt []*packet, mtu MTU, env Environment) []*packet {
	nPkt := NewPacket(*GetPktsDest(pkt), *GetPktsSrc(pkt), ICMP_REP, DefragmentData(pkt), 8, 0, 0)
	return Fragment(&nPkt, mtu)
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

	icmpReqPkt := createIcmpReq(n.name, dstName, n.netPort, dstNetPort, msg)
	pkt := Fragment(&icmpReqPkt, dstNetPort.mtu)
	env.SendIcmpReq(n, n.netPort, dstNetPort, pkt)
}

/*
----------------------------------------------------
Router implementations
----------------------------------------------------
*/

type routerTableEntry struct {
	netdest IP
	nexthop IP
	port    uint8
}

type routerPort struct {
	number uint8
	netInterface
}

// NewRouterPort function creates a new port for a router
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

type Router interface {
	SendIcmpTimeExceeded(pkts []*packet, env Environment) []*packet
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

/*
----------------------------------------------------
Basic functions for the router
----------------------------------------------------
*/

// NewRouter creates a new router
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

/*
----------------------------------------------------
Router send functions implementation
----------------------------------------------------
*/

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

func (r *router) SendIcmpReply(pkt []*packet, mtu MTU, env Environment) []*packet {
	// verify if the router can reach the network
	entry, _ := fp.Find(r.routerTable, func(ent *routerTableEntry) bool {
		return ent.netdest.IsSameNet(GetPktsDest(pkt).ip)
	})

	rtEntry := entry.(*routerTableEntry)
	defaultIp := *NewIp("0.0.0.0/0")

	var destNetInterface netInterface
	var routerNetPort netInterface
	var destNetComp NetComponent

	// retrieve the port that can reach the netork
	port, _ := fp.Find(r.ports, func(p routerPort) bool {
		return p.number == rtEntry.port
	})
	routerNetPort = port.(routerPort).netInterface

	if rtEntry.nexthop == defaultIp {
		destNetComp = env.GetNetComponentByIp(GetPktsDest(pkt).ip)
		destNetInterface = env.GetComponentNetInterfaceByIp(destNetComp, GetPktsDest(pkt).ip)
	} else {
		destNetComp = env.GetNetComponentByIpOnly(rtEntry.nexthop)
		destNetInterface = env.GetComponentNetInterfaceByIpOnly(destNetComp, rtEntry.nexthop)
	}

	// verify if the destination is known by the router
	_, hasMacArpTable := r.arpTable[destNetInterface.ip]
	if !hasMacArpTable {
		pkt := createBroadcastArpReq(r.name, routerNetPort, destNetInterface.ip)
		arpReply := env.SendArpReq(pkt)
		r.ReceiveArpRequest(arpReply)
	}

	srcHost := &packetHost{
		name: r.name,
		ip:   GetPktsSrc(pkt).ip,
		mac:  routerNetPort.mac,
	}
	dstHost := &packetHost{
		ip:   GetPktsDest(pkt).ip,
		mac:  destNetInterface.mac,
		name: destNetComp.GetName(),
	}

	list, _ := fp.Map(pkt, func(pkt *packet) []*packet {
		return Fragment(pkt, destNetInterface.mtu)
	})
	pktList := list.([][]*packet)

	var pkts []*packet = make([]*packet, 0)
	for _, p := range pktList {
		pkts = append(pkts, p...)
	}

	SetHosts(pkts, srcHost, dstHost)
	return pkts
}

func (r *router) SendIcmpTimeExceeded(pkt []*packet, env Environment) []*packet {
	if GetPktsType(pkt) == ICMP_TIME_EXCEEDED {
		return pkt
	}

	// verify if the router can reach the network
	entry, _ := fp.Find(r.routerTable, func(ent *routerTableEntry) bool {
		return ent.netdest.IsSameNet(GetPktsSrc(pkt).ip)
	})

	rtEntry := entry.(*routerTableEntry)
	defaultIp := *NewIp("0.0.0.0/0")

	var destNetInterface netInterface
	var routerNetPort netInterface
	var destNetComp NetComponent

	// retrieve the port that can reach the netork
	port, _ := fp.Find(r.ports, func(p routerPort) bool {
		return p.number == rtEntry.port
	})
	routerNetPort = port.(routerPort).netInterface

	if rtEntry.nexthop == defaultIp {
		destNetComp = env.GetNetComponentByIp(GetPktsSrc(pkt).ip)
		destNetInterface = env.GetComponentNetInterfaceByIp(destNetComp, GetPktsSrc(pkt).ip)
	} else {
		destNetComp = env.GetNetComponentByIpOnly(rtEntry.nexthop)
		destNetInterface = env.GetComponentNetInterfaceByIpOnly(destNetComp, rtEntry.nexthop)
	}

	// verify if the destination is known by the router
	_, hasMacArpTable := r.arpTable[destNetInterface.ip]
	if !hasMacArpTable {
		pkt := createBroadcastArpReq(r.name, routerNetPort, destNetInterface.ip)
		arpReply := env.SendArpReq(pkt)
		r.ReceiveArpRequest(arpReply)
	}

	srcHost := packetHost{
		name: r.name,
		ip:   routerNetPort.ip,
		mac:  routerNetPort.mac,
	}
	dstHost := packetHost{
		name: destNetComp.GetName(),
		ip:   GetPktsSrc(pkt).ip,
		mac:  destNetInterface.mac,
	}
	timePkt := NewPacket(srcHost, dstHost, ICMP_TIME_EXCEEDED, DefragmentData(pkt), 8, 0, 0)
	return Fragment(&timePkt, destNetInterface.mtu)
}

/*
----------------------------------------------------
Router receive functions implementations
----------------------------------------------------
*/

func (r *router) ReceiveArpRequest(pkt packet) {
	_, hasSrcIp := r.arpTable[pkt.src.ip]

	if !hasSrcIp {
		r.arpTable[pkt.src.ip] = pkt.src.mac
	}
}

func (r *router) ReceiveIcmpRequest(pkt []*packet, env Environment) bool {
	// verify if the router can reach the network
	entry, _ := fp.Find(r.routerTable, func(ent *routerTableEntry) bool {
		return ent.netdest.IsSameNet(GetPktsDest(pkt).ip)
	})

	rtEntry := entry.(*routerTableEntry)
	defaultIp := *NewIp("0.0.0.0/0")

	var destNetInterface netInterface
	var routerNetPort netInterface
	var destNetComp NetComponent

	// retrieve the port that can reach the netork
	port, _ := fp.Find(r.ports, func(p routerPort) bool {
		return p.number == rtEntry.port
	})
	routerNetPort = port.(routerPort).netInterface

	if rtEntry.nexthop == defaultIp {
		destNetComp = env.GetNetComponentByIp(GetPktsDest(pkt).ip)
		destNetInterface = env.GetComponentNetInterfaceByIp(destNetComp, GetPktsDest(pkt).ip)
	} else {
		destNetComp = env.GetNetComponentByIpOnly(rtEntry.nexthop)
		destNetInterface = env.GetComponentNetInterfaceByIpOnly(destNetComp, rtEntry.nexthop)
	}

	// verify if the destination is known by the router
	_, hasMacArpTable := r.arpTable[destNetInterface.ip]
	if !hasMacArpTable {
		pkt := createBroadcastArpReq(r.name, routerNetPort, destNetInterface.ip)
		arpReply := env.SendArpReq(pkt)
		r.ReceiveArpRequest(arpReply)
	}

	srcNetInterface := netInterface{
		ip:  GetPktsSrc(pkt).ip,
		mac: routerNetPort.mac,
		mtu: routerNetPort.mtu,
	}
	destNetInterface = netInterface{
		ip:  GetPktsDest(pkt).ip,
		mac: destNetInterface.mac,
		mtu: destNetInterface.mtu,
	}
	srcHost := &packetHost{
		name: r.name,
		ip:   GetPktsSrc(pkt).ip,
		mac:  routerNetPort.mac,
	}
	dstHost := &packetHost{
		name: destNetComp.GetName(),
		ip:   GetPktsDest(pkt).ip,
		mac:  destNetInterface.mac,
	}

	DecrementPktsTTL(pkt)
	SetHosts(pkt, srcHost, dstHost)
	env.SendIcmpReq(r, srcNetInterface, destNetInterface, pkt)
	return false
}

func (r *router) ReceiveIcmpReply(pkts []*packet, env Environment) {
	// verify if the router can reach the network
	entry, _ := fp.Find(r.routerTable, func(ent *routerTableEntry) bool {
		return ent.netdest.IsSameNet(GetPktsDest(pkts).ip)
	})

	rtEntry := entry.(*routerTableEntry)
	defaultIp := *NewIp("0.0.0.0/0")

	var destNetInterface netInterface
	var routerNetPort netInterface
	var destNetComp NetComponent

	// retrieve the port that can reach the netork
	port, _ := fp.Find(r.ports, func(p routerPort) bool {
		return p.number == rtEntry.port
	})
	routerNetPort = port.(routerPort).netInterface

	if rtEntry.nexthop == defaultIp {
		destNetComp = env.GetNetComponentByIp(GetPktsDest(pkts).ip)
		destNetInterface = env.GetComponentNetInterfaceByIp(destNetComp, GetPktsDest(pkts).ip)
	} else {
		destNetComp = env.GetNetComponentByIpOnly(rtEntry.nexthop)
		destNetInterface = env.GetComponentNetInterfaceByIpOnly(destNetComp, rtEntry.nexthop)
	}

	// verify if the destination is known by the router
	_, hasMacArpTable := r.arpTable[destNetInterface.ip]
	if !hasMacArpTable {
		pkt := createBroadcastArpReq(r.name, routerNetPort, destNetInterface.ip)
		arpReply := env.SendArpReq(pkt)
		r.ReceiveArpRequest(arpReply)
	}

	srcHost := &packetHost{
		name: r.name,
		ip:   GetPktsSrc(pkts).ip,
		mac:  routerNetPort.mac,
	}
	destHost := &packetHost{
		name: destNetComp.GetName(),
		ip:   GetPktsDest(pkts).ip,
		mac:  destNetInterface.mac,
	}

	list, _ := fp.Map(pkts, func(pkt *packet) []*packet {
		frag := Fragment(pkt, destNetInterface.mtu)
		return frag
	})
	pktList := list.([][]*packet)

	var pktsToGo []*packet = make([]*packet, 0)
	for _, p := range pktList {
		pktsToGo = append(pktsToGo, p...)
	}
	SetHosts(pktsToGo, srcHost, destHost)
	DecrementPktsTTL(pktsToGo)
	env.SendIcmpReply(r, destNetComp, pktsToGo)
}

func (r *router) ReceiveTimeExceeded(pkt []*packet, env Environment) {
	// verify if the router can reach the network
	entry, _ := fp.Find(r.routerTable, func(ent *routerTableEntry) bool {
		return ent.netdest.IsSameNet(GetPktsDest(pkt).ip)
	})

	rtEntry := entry.(*routerTableEntry)
	defaultIp := *NewIp("0.0.0.0/0")

	var destNetInterface netInterface
	var routerNetPort netInterface
	var destNetComp NetComponent

	// retrieve the port that can reach the netork
	port, _ := fp.Find(r.ports, func(p routerPort) bool {
		return p.number == rtEntry.port
	})
	routerNetPort = port.(routerPort).netInterface

	if rtEntry.nexthop == defaultIp {
		destNetComp = env.GetNetComponentByIp(GetPktsDest(pkt).ip)
		destNetInterface = env.GetComponentNetInterfaceByIp(destNetComp, GetPktsDest(pkt).ip)
	} else {
		destNetComp = env.GetNetComponentByIpOnly(rtEntry.nexthop)
		destNetInterface = env.GetComponentNetInterfaceByIpOnly(destNetComp, rtEntry.nexthop)
	}

	// verify if the destination is known by the router
	_, hasMacArpTable := r.arpTable[destNetInterface.ip]
	if !hasMacArpTable {
		pkt := createBroadcastArpReq(r.name, routerNetPort, destNetInterface.ip)
		arpReply := env.SendArpReq(pkt)
		r.ReceiveArpRequest(arpReply)
	}

	srcHost := &packetHost{
		name: r.name,
		ip:   GetPktsSrc(pkt).ip,
		mac:  routerNetPort.mac,
	}
	destHost := &packetHost{
		name: destNetComp.GetName(),
		ip:   GetPktsDest(pkt).ip,
		mac:  destNetInterface.mac,
	}
	list, _ := fp.Map(pkt, func(pkt *packet) []*packet {
		return Fragment(pkt, destNetInterface.mtu)
	})
	pktList := list.([][]*packet)

	var pkts []*packet = make([]*packet, 0)
	for _, p := range pktList {
		pkts = append(pkts, p...)
	}
	DecrementPktsTTL(pkts)
	SetHosts(pkts, srcHost, destHost)
	env.SendIcmpTimeExceeded(r, pkts)
}

/*
----------------------------------------------------
Environment implementations
----------------------------------------------------
*/

type Environment interface {
	AddNode(nd *node)
	AddRouter(r *router)
	GetDefaultGateway(n *node) *router
	GetNetComponentByName(name string) NetComponent
	GetNetComponentByIp(ip IP) NetComponent
	GetNetComponentByIpOnly(ip IP) NetComponent
	GetComponentNetInterfaceByIp(comp NetComponent, ip IP) netInterface
	GetComponentNetInterfaceByIpOnly(comp NetComponent, ip IP) netInterface
	ParseLines(lines []string)

	SendMessage(msg string, ipSrc, ipDest IP) error
	SendArpReq(pkt packet) packet
	SendIcmpReq(srcComp NetComponent, srcNetPort, dstNetPort netInterface, pkts []*packet)
	SendIcmpReply(src, dest NetComponent, pkts []*packet)
	SendIcmpTimeExceeded(src NetComponent, pkts []*packet)
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

func (e *environment) GetNetComponentByIpOnly(ip IP) NetComponent {
	var component NetComponent
	for _, r := range e.routers {
		for _, rp := range r.ports {
			if rp.ip.ip == ip.ip {
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
	logArpRequest(pkt)
	dst := e.GetNetComponentByIp(pkt.dst.ip)
	dst.ReceiveArpRequest(pkt)
	arpReply := dst.SendArpReply(pkt)
	logArpReply(arpReply)
	return arpReply
}

func (e *environment) SendIcmpReq(src NetComponent, srcNetPort, dstNetPort netInterface, pkts []*packet) {
	if IsTimeExceeded(pkts) {
		e.SendIcmpTimeExceeded(src, pkts)
		return
	}
	logIcmpRequest(pkts)

	dst := e.GetNetComponentByMac(GetPktsDest(pkts).mac)

	doReply := dst.ReceiveIcmpRequest(pkts, e)
	if doReply {
		e.SendIcmpReply(dst, src, pkts)
	}
}

func (e *environment) SendIcmpReply(src, dest NetComponent, pkts []*packet) {
	var mtu MTU

	switch src.(type) {
	case *node:
		mtu = src.(*node).GetNetInterface().mtu
	default:
		mtu = 255
	}

	replyPkts := src.SendIcmpReply(pkts, mtu, e)

	if IsTimeExceeded(replyPkts) {
		e.SendIcmpTimeExceeded(src, pkts)
		return
	}
	logIcmpReply(replyPkts)

	destination := e.GetNetComponentByMac(GetPktsDest(replyPkts).mac)
	destination.ReceiveIcmpReply(replyPkts, e)
}

func (e *environment) SendIcmpTimeExceeded(src NetComponent, pkt []*packet) {
	timePkt := src.(Router).SendIcmpTimeExceeded(pkt, e)
	logIcmpTimeExceeded(timePkt)
	destination := e.GetNetComponentByMac(GetPktsDest(timePkt).mac)
	destination.ReceiveTimeExceeded(timePkt, e)
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

func (e *environment) GetComponentNetInterfaceByIpOnly(comp NetComponent, ip IP) netInterface {
	switch comp := comp.(type) {
	case *node:
		return comp.GetNetInterface()
	case *router:
		for _, port := range comp.ports {
			if port.ip.ip == ip.ip {
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
	src.(Node).SendMessage(msg, dst, destNetInterface, e)
	return nil
}

/*
----------------------------------------------------
Parse the file to create the environment
----------------------------------------------------
*/

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

/*
----------------------------------------------------
Run the Simulator
----------------------------------------------------
*/

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

	dest := env.GetNetComponentByName(args.DstNode)
	ipDest := dest.(Node).GetNetInterface()

	return env.SendMessage(args.Msg, ipSrc.ip, ipDest.ip)
}
