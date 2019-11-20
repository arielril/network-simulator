package file

import (
	"errors"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/urfave/cli"
)

type InputArgs struct {
	Topology string `valid:"alphanum"`
	SrcNode  string `valid:"alphanum"`
	DstNode  string `valid:"alphanumn"`
	Msg      string `valid:"ascii"`
}

func ValidateInputeArgs(args *InputArgs, ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return errors.New("Invalid simulator arguments")
	}

	args.Topology = ctx.Args().Get(0)
	args.SrcNode = ctx.Args().Get(1)
	args.DstNode = ctx.Args().Get(2)
	args.Msg = ctx.Args().Get(3)

	_, err := govalidator.ValidateStruct(args)

	if err != nil {
		return errors.New(
			fmt.Sprintf("Falied to parse args: %v", err),
		)
	}

	return nil
}
