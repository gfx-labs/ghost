package cli

import (
	"fmt"
	"io"
	"os"

	"gfx.cafe/open/ghost"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func NewCli(g ghost.Client) *cli.App {
	app := &cli.App{
		Name:  "spook",
		Usage: "cli for chain",
		ExitErrHandler: func(cCtx *cli.Context, err error) {
			if err != nil {
				fmt.Print(err)
			}
		},
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
			{
				Name:      "disasm",
				Usage:     "disassemble bytes. use - for input from stdin",
				UsageText: "spook disasm [file]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "swarm",
						DefaultText: "whether or not to deal with the swarm hash",
					},
					&cli.BoolFlag{
						Name:        "constructor",
						DefaultText: "whether or not the constructor is included in input",
						Aliases:     []string{"ctor"},
					},
					&cli.BoolFlag{
						Name:        "logging",
						DefaultText: "enable logging",
						Aliases:     []string{"log"},
					},
					&cli.BoolFlag{
						Name:        "binary",
						DefaultText: "binary input",
						Aliases:     []string{"bin"},
					},
				},
				Action: func(ctx *cli.Context) error {
					file := ctx.Args().Get(0)
					if file == "" {
						return cli.ShowSubcommandHelp(ctx)
					}
					var in io.Reader
					if file == "-" {
						in = os.Stdin
					} else {
						a, err := os.Open(file)
						if err != nil {
							return err
						}
						defer a.Close()
						in = a
					}
					return Disasm(
						in,
						ctx.Bool("swarm"),
						ctx.Bool("constructor"),
						ctx.Bool("logging"),
						ctx.Bool("binary"),
					)
				},
			},
		},
	}
	return app
}
