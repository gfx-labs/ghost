package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "ghost",
		Usage: "EVM ABI toolkit — encode, decode, call, and navigate ABI data",
		Commands: []*cli.Command{
			sigCmd,
			encodeCmd,
			decodeCmd,
			pathCmd,
			callCmd,
			txCmd,
			balanceCmd,
			blockCmd,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "rpc",
				Aliases: []string{"r"},
				Usage:   "JSON-RPC endpoint URL",
				Sources: cli.EnvVars("ETH_RPC_URL"),
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		log.SetFlags(0)
		os.Exit(1)
	}
}
