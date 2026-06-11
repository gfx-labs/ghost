package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/gfx-labs/ghost/abi"
	"github.com/urfave/cli/v3"
)

var sigCmd = &cli.Command{
	Name:      "sig",
	Usage:     "compute function selector and keccak256 hash of a signature",
	ArgsUsage: "<signature>",
	Description: `Compute the 4-byte selector and full keccak256 hash of a function signature.

Examples:
  ghost sig "transfer(address,uint256)"
  ghost sig "balanceOf(address)"
  ghost sig "aggregate3((address,bool,bytes)[])"`,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() < 1 {
			return fmt.Errorf("usage: ghost sig <signature>")
		}
		sig := abi.Signature(cmd.Args().First())
		sel := sig.Selector()
		hash := sig.Hash()
		fmt.Printf("signature:  %s\n", sig)
		fmt.Printf("selector:   0x%s\n", hex.EncodeToString(sel[:]))
		fmt.Printf("hash:       0x%s\n", hex.EncodeToString(hash[:]))
		return nil
	},
}
