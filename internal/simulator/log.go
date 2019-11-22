package simulator

import "fmt"

func logArpRequest(pkt packet) {
	fmt.Printf(
		"%v box %v : ETH (src=%v dst=%v) \\n ARP - Who has %v? Tell %v;\n",
		pkt.src.name, pkt.src.name, pkt.src.mac, pkt.dst.mac,
		pkt.dst.ip.ip, pkt.src.ip.ip,
	)
}

func logArpReply(pkt packet) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \\n ARP - %v is at %v;\n",
		pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
		pkt.src.ip.ip, pkt.src.mac,
	)
}

func logIcmpRequest(pkt packet) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Echo request (data=%v);\n",
		pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
		pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off, pkt.data,
	)
}

func logIcmpReply(pkt packet) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Echo reply (data=%v);\n",
		pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
		pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off, pkt.data,
	)
}

func logIcmpTimeExceeded(pkt packet) {
	fmt.Printf(
		"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Time Exceeded;\n",
		pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
		pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off,
	)
}
