package main

import (
	"fmt"
	"os"

	"github.com/arielril/network-simulator/internal/simulator"

	"github.com/urfave/cli"
)

func configCli(app *cli.App) {
	app.Name = "Network Simulator"
	app.Usage = "Let's you run a ping simulation inside a topology"
}

func getCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Network Simulator"
	app.Usage = "Let's you run a ping simulation inside a topology"
	app.UsageText = "simulador [path/to/topology/file] [src_node] [dst_node] [message]"
	app.Action = simulator.Run

	return app
}

func main() {
	app := getCliApp()
	appErr := app.Run(os.Args)

	if appErr != nil {
		panic(fmt.Sprintf("Failed to run cli: %v", appErr))
	}
}
