package simulator

import (
	"fmt"
	"path/filepath"

	"github.com/arielril/network-simulator/internal/network"

	"github.com/arielril/network-simulator/internal/component"

	"github.com/arielril/network-simulator/internal/file"
	"github.com/urfave/cli"
)

var (
	ip1 string = "192.168.0.2"
	ip2 string = "192.168.1.3"
)

func Run(ctx *cli.Context) error {
	topologyFile := ctx.Args().First()

	filePath, _ := filepath.Abs(topologyFile)
	fileR := file.Read(filePath)

	net := network.CreateNetwork(fileR)
	fmt.Println(net.Nodes)
	fmt.Println(net.Routers)

	srcNode := component.Node{
		Name: ctx.Args().Get(1),
	}
	destNode := component.Node{
		Name: ctx.Args().Get(2),
	}
	message := ctx.Args().Get(3)

	net.SendMsg(srcNode, destNode, message)

	return nil
}
