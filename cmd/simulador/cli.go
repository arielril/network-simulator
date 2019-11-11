package main

import (
	"fmt"

	"github.com/arielril/network-simulator/internal/file"

	"github.com/arielril/network-simulator/internal/component"
)

var (
	ip1 string = "192.168.0.2"
	ip2 string = "192.168.1.3"
)

func main() {
	fmt.Println("Hey!")

	ipSrc := component.IP{Ip: ip1, Prefix: 24}
	ipDest := component.IP{Ip: ip2, Prefix: 24}
	fmt.Printf(
		"Is %v/%v in the same net as %v/%v? %v\n",
		ipSrc.Ip,
		ipSrc.Prefix,
		ipDest.Ip,
		ipDest.Prefix,
		ipSrc.IsSameNet(ipDest),
	)
	fileR := file.Read("/Users/arielril/Documents/go.nosync/src/github.com/arielril/network-simulator/examples/example1.txt")
	file.Parse(fileR)
}
