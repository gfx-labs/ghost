package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gfx-labs/ghost/abi"
	"github.com/gfx-labs/ghost/abipath"
	"github.com/urfave/cli/v3"
)

var callCmd = &cli.Command{
	Name:      "call",
	Usage:     "perform an eth_call and optionally decode/navigate the result",
	ArgsUsage: `<address> <signature> [args...]`,
	Description: `Execute an eth_call against a contract and display the result.
Combine with --path to navigate the return data using abipath,
or --decode to decode the return data with type hints.

Examples:
  ghost call 0xtoken "balanceOf(address)" 0xuser --rpc $RPC
  ghost call 0xtoken "totalSupply()" --rpc $RPC --decode "uint256"
  ghost call 0xtoken "name()" --rpc $RPC --decode "string"
  ghost call 0xpair "getReserves()" --rpc $RPC --path ".0"
  ghost call 0xaddr "getData()" --rpc $RPC --path ".0.1" --decode "uint256"`,
	Flags: []cli.Flag{
		outputFlag,
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "abipath expression to navigate the return data",
		},
		&cli.StringFlag{
			Name:    "decode",
			Aliases: []string{"d"},
			Usage:   "type descriptor to decode the result (e.g. 'uint256', '(uint256,address)')",
		},
		&cli.StringFlag{
			Name:  "block",
			Value: "latest",
			Usage: "block number or tag (latest, pending, earliest)",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() < 2 {
			return fmt.Errorf("usage: ghost call <address> <signature> [args...]")
		}

		rpcURL, err := getRPC(cmd)
		if err != nil {
			return err
		}

		addr := cmd.Args().Get(0)
		if !strings.HasPrefix(addr, "0x") {
			addr = "0x" + addr
		}

		sigStr := cmd.Args().Get(1)
		sig := abi.Signature(sigStr)
		args := sig.Args().TupleArgs()

		providedArgs := cmd.NArg() - 2
		if providedArgs != len(args) {
			return fmt.Errorf("expected %d arguments for %s, got %d", len(args), sigStr, providedArgs)
		}

		// Build calldata
		b := &abi.Builder{}
		for i, t := range args {
			argStr := cmd.Args().Get(i + 2)
			if err := encodeArg(b, t, argStr); err != nil {
				return fmt.Errorf("argument %d (%s): %w", i, t, err)
			}
		}
		calldata := b.Finish(sig.Fn())

		// Make eth_call
		callObj := map[string]string{
			"to":   addr,
			"data": "0x" + hex.EncodeToString(calldata),
		}
		result, err := rpcCall(rpcURL, "eth_call", callObj, cmd.String("block"))
		if err != nil {
			return err
		}

		// Parse hex result
		var resultHex string
		if err := json.Unmarshal(result, &resultHex); err != nil {
			return fmt.Errorf("parsing result: %w", err)
		}

		data, err := parseHex(resultHex)
		if err != nil {
			return fmt.Errorf("parsing result hex: %w", err)
		}

		// Apply abipath if specified
		if p := cmd.String("path"); p != "" {
			data, err = abipath.Point(p, data)
			if err != nil {
				return fmt.Errorf("path navigation: %w", err)
			}
		}

		// Decode if type hint given
		if d := cmd.String("decode"); d != "" {
			types := abi.TypeName(d)
			if types.IsTuple() {
				targs := types.TupleArgs()
				dec := abi.NewDecoder(data)
				for i, t := range targs {
					val, err := decodeAndFormat(dec, t)
					if err != nil {
						return fmt.Errorf("decoding arg %d: %w", i, err)
					}
					fmt.Printf("[%d] %s: %s\n", i, t, val)
				}
				return nil
			}
			dec := abi.NewDecoder(data)
			val, err := decodeAndFormat(dec, types)
			if err != nil {
				return fmt.Errorf("decoding: %w", err)
			}
			return writeString(cmd, val)
		}

		// Raw hex output
		return writeOutput(cmd, data)
	},
}
