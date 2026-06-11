package main

import (
	"context"
	"fmt"

	"github.com/gfx-labs/ghost/abipath"
	"github.com/urfave/cli/v3"
)

var pathCmd = &cli.Command{
	Name:      "path",
	Usage:     "navigate ABI-encoded data using abipath expressions",
	ArgsUsage: `<path> [hex]`,
	Description: `Navigate ABI-encoded data using abipath path expressions.
Hex data can be passed as an argument, piped via stdin, or read from a file.

Path syntax:
  .    follow a dynamic offset (reads uint256 offset, jumps to that position)
  /    follow a dynamic offset and skip the length prefix
  N    skip N 32-byte words

Examples:
  ghost path ".0" 0x000...       # follow first dynamic offset, read from there
  ghost path ".1" 0x000...       # follow dynamic, skip 1 word
  ghost path "/2" 0x000...       # follow dynamic+length, skip 2 words
  echo "0x..." | ghost path ".0" # pipe from stdin
  ghost call 0xaddr "fn()" --rpc $RPC | ghost path ".0"  # chain with call`,
	Flags: []cli.Flag{
		inputFlag, outputFlag,
		&cli.BoolFlag{
			Name:    "selector",
			Aliases: []string{"s"},
			Usage:   "strip the 4-byte function selector before navigating",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() < 1 {
			return fmt.Errorf("usage: ghost path <path-expr> [hex]")
		}

		pathExpr := cmd.Args().First()
		data, err := readHexInput(cmd, 1)
		if err != nil {
			return err
		}

		if cmd.Bool("selector") {
			if len(data) < 4 {
				return fmt.Errorf("data too short to contain a selector")
			}
			data = data[4:]
		}

		result, err := abipath.Point(pathExpr, data)
		if err != nil {
			return fmt.Errorf("path navigation failed: %w", err)
		}

		return writeOutput(cmd, result)
	},
}
