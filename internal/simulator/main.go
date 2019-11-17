package simulator

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/arielril/network-simulator/internal/network"
	"github.com/asaskevich/govalidator"

	"github.com/arielril/network-simulator/internal/component"

	"github.com/arielril/network-simulator/internal/file"
	fp "github.com/novalagung/gubrak"
	"github.com/urfave/cli"
)

var (
	ip1 string = "192.168.0.2"
	ip2 string = "192.168.1.3"
)

type InputArgs struct {
	topology string `valid:"alphanum"`
	srcNode  string `valid:"alphanum"`
	dstNode  string `valid:"alphanumn"`
	msg      string `valid:"ascii"`
}

func validateInputeArgs(args *InputArgs, ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return errors.New("Invalid simulator arguments")
	}

	args.topology = ctx.Args().Get(0)
	args.srcNode = ctx.Args().Get(1)
	args.dstNode = ctx.Args().Get(2)
	args.msg = ctx.Args().Get(3)

	_, err := govalidator.ValidateStruct(args)

	if err != nil {
		return errors.New(
			fmt.Sprintf("Falied to parse args: %v", err),
		)
	}

	return nil
}

func findNodeByName(nodes []component.Node) func(string) component.Node {
	return func(name string) component.Node {
		nd, _ := fp.Find(nodes, func(node component.Node) bool {
			return node.Name == name
		})

		return nd.(component.Node)
	}
}

func Run(ctx *cli.Context) error {
	args := &InputArgs{}

	validateInputeArgs(args, ctx)

	filePath, _ := filepath.Abs(args.topology)
	fileR := file.Read(filePath)
	net := network.CreateNetwork(fileR)

	findNode := findNodeByName(net.Nodes)

	srcNode := findNode(args.srcNode)
	dstNode := findNode(args.srcNode)

	message := args.msg

	net.SendMsg(srcNode, dstNode, message)

	return nil
}
