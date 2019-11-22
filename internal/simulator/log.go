package simulator

import (
	"fmt"

	fp "github.com/novalagung/gubrak"
)

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

func logIcmpRequest(pkts []*packet) {
	fp.ForEach(pkts, func(pkt *packet) {
		fmt.Printf(
			"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Echo request (data=%v);\n",
			pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
			pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off, pkt.data,
		)
	})
}

func logIcmpReply(pkts []*packet) {
	fp.ForEach(pkts, func(pkt *packet) {
		fmt.Printf(
			"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Echo reply (data=%v);\n",
			pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
			pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off, pkt.data,
		)
	})
}

func logIcmpTimeExceeded(pkts []*packet) {
	fp.ForEach(pkts, func(pkt *packet) {
		fmt.Printf(
			"%v => %v : ETH (src=%v dst=%v) \\n IP (src=%v dst=%v ttl=%v mf=%v off=%v) \\n ICMP - Time Exceeded;\n",
			pkt.src.name, pkt.dst.name, pkt.src.mac, pkt.dst.mac,
			pkt.src.ip.ip, pkt.dst.ip.ip, pkt.ttl, pkt.mf, pkt.off,
		)
	})
}
