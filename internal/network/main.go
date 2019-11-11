package network

import (
	"fmt"

	. "github.com/arielril/network-simulator/internal/component"
)

type Network struct {
	nodes   []Node
	routers []Router
}

type INetwork interface {
	SendMsg(srcNode Node, dstNode Node, msg string)
}

func (net Network) SendMsg(srcNode Node, dstNode Node, msg string) {
	fmt.Printf("From: %v\n", srcNode.Name)
	fmt.Printf("To: %v\n", dstNode.Name)
	fmt.Printf("   Message: %v\n", msg)
}
