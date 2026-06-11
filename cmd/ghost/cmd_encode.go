package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gfx-labs/ghost/abi"
	"github.com/urfave/cli/v3"
)

var encodeCmd = &cli.Command{
	Name:      "encode",
	Usage:     "ABI-encode calldata from a function signature and arguments",
	ArgsUsage: `<signature> [args...]`,
	Description: `Encode function calldata from a human-readable signature and arguments.

Examples:
  ghost encode "transfer(address,uint256)" 0xdead...beef 1000000
  ghost encode "baz(uint32,bool)" 69 true
  ghost encode "sam(bytes,bool,uint256[])" 0x64617665 true "[1,2,3]"

Types are inferred from the signature. Supported argument formats:
  - integers: decimal or 0x hex
  - bool: true/false
  - address: 0x-prefixed hex
  - bytes: 0x-prefixed hex
  - string: plain text (use quotes in shell)
  - arrays: [val1,val2,val3]`,
	Flags: []cli.Flag{outputFlag},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() < 1 {
			return fmt.Errorf("usage: ghost encode <signature> [args...]")
		}

		sigStr := cmd.Args().First()
		sig := abi.Signature(sigStr)
		args := sig.Args().TupleArgs()

		if cmd.NArg()-1 != len(args) {
			return fmt.Errorf("expected %d arguments for %s, got %d", len(args), sigStr, cmd.NArg()-1)
		}

		b := &abi.Builder{}
		for i, t := range args {
			argStr := cmd.Args().Get(i + 1)
			if err := encodeArg(b, t, argStr); err != nil {
				return fmt.Errorf("argument %d (%s): %w", i, t, err)
			}
		}

		data := b.Finish(sig.Fn())
		return writeOutput(cmd, data)
	},
}

func encodeArg(b *abi.Builder, t abi.TypeName, s string) error {
	st := string(t)

	switch {
	// Composite types first (before prefix matches eat uint256[] etc.)
	case t.IsSlice():
		elem, _ := t.UnSlice()
		items := parseArrayArg(s)
		arr := b.EnterDynamicArray()
		for _, item := range items {
			if err := encodeArg(arr, elem, item); err != nil {
				return err
			}
		}
		arr.Exit()
		return nil

	case t.IsTuple():
		return fmt.Errorf("inline tuple encoding not yet supported; use explicit calldata")

	case t == abi.BOOL:
		switch strings.ToLower(s) {
		case "true", "1":
			b.Bool(true)
		case "false", "0":
			b.Bool(false)
		default:
			return fmt.Errorf("invalid bool: %q", s)
		}
		return nil

	case t == abi.ADDRESS:
		if !strings.HasPrefix(s, "0x") && !strings.HasPrefix(s, "0X") {
			s = "0x" + s
		}
		b.Address(common.HexToAddress(s))
		return nil

	case t == abi.STRING:
		b.DString(s)
		return nil

	case t == abi.BYTES:
		bs, err := parseHex(s)
		if err != nil {
			return err
		}
		b.Bytes(bs)
		return nil

	case strings.HasPrefix(st, "bytes"):
		bs, err := parseHex(s)
		if err != nil {
			return err
		}
		nStr := strings.TrimPrefix(st, "bytes")
		var n int
		fmt.Sscanf(nStr, "%d", &n)
		b.FixedBytes(n, bs)
		return nil

	case strings.HasPrefix(st, "uint") || strings.HasPrefix(st, "int") ||
		strings.HasPrefix(st, "fixed") || strings.HasPrefix(st, "ufixed"):
		v, ok := new(big.Int).SetString(s, 0)
		if !ok {
			return fmt.Errorf("invalid number: %q", s)
		}
		b.BigInt(v)
		return nil

	default:
		return fmt.Errorf("unsupported type: %s", t)
	}
}

// parseArrayArg parses "[1,2,3]" into []string{"1","2","3"}
func parseArrayArg(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
