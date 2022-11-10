package cli

import (
	"gfx.cafe/open/ghost"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func NewCli(g ghost.Client) *cli.App {
	app := &cli.App{
		Name:  "spook",
		Usage: "cli for chain",
		Commands: []*cli.Command{
			{
				Name:      "contract",
				Aliases:   []string{"c"},
				Usage:     "interact with contract",
				ArgsUsage: "[address]",
				Action: func(ctx *cli.Context) error {
					if ctx.Args().Len() == 0 {
						cli.ShowCommandHelp(ctx, "contract")
						return nil
					}
					ci := NewContract(g, "?", common.HexToAddress(ctx.Args().First()))
					nestedArgs := ctx.Args().Slice()[1:]
					ci.Run(nestedArgs)
					return nil
				},
			},
		},
	}
	return app
}
