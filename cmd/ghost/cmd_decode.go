package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gfx-labs/ghost/abi"
	"github.com/urfave/cli/v3"
)

var decodeCmd = &cli.Command{
	Name:      "decode",
	Usage:     "ABI-decode hex data given type descriptors",
	ArgsUsage: `<types> [hex]`,
	Description: `Decode ABI-encoded data given a type signature or tuple descriptor.
Hex data can be passed as an argument, piped via stdin, or read from a file.

Examples:
  ghost decode "(uint256,address)" 0x000...
  ghost decode "uint256" 0x000...
  echo "0x000..." | ghost decode "(uint256,string)"
  ghost decode "(uint256,string)" -i data.hex

If the type starts with a function name (e.g. "transfer(address,uint256)"),
the first 4 bytes are treated as the selector and skipped.`,
	Flags: []cli.Flag{inputFlag, outputFlag},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() < 1 {
			return fmt.Errorf("usage: ghost decode <types> [hex]")
		}

		typeStr := cmd.Args().First()
		data, err := readHexInput(cmd, 1)
		if err != nil {
			return err
		}

		// Detect if this is a full function signature (has a name before the parens)
		var types abi.TypeName
		if i := strings.Index(typeStr, "("); i > 0 && typeStr[0] != '(' {
			// Function signature — skip the 4-byte selector
			if len(data) < 4 {
				return fmt.Errorf("calldata too short to contain a selector")
			}
			sig := abi.Signature(typeStr)
			types = sig.Args()
			data = data[4:]
		} else {
			types = abi.TypeName(typeStr)
		}

		// If it's a tuple, decode each element
		if types.IsTuple() {
			args := types.TupleArgs()
			dec := abi.NewDecoder(data)
			for i, t := range args {
				val, err := decodeAndFormat(dec, t)
				if err != nil {
					return fmt.Errorf("arg %d (%s): %w", i, t, err)
				}
				fmt.Printf("[%d] %s: %s\n", i, t, val)
			}
			return nil
		}

		// Single type
		dec := abi.NewDecoder(data)
		val, err := decodeAndFormat(dec, types)
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

// decodeAndFormat reads a single ABI-typed value from the decoder and returns
// a human-readable string. Composite types (slices, tuples) are checked first
// so that e.g. "uint256[]" doesn't match the "uint" prefix case.
func decodeAndFormat(dec *abi.Decoder, t abi.TypeName) (string, error) {
	st := string(t)

	switch {
	// --- composite types first ---
	case t.IsSlice():
		elem, _ := t.UnSlice()
		sub, l, err := dec.DynamicLength()
		if err != nil {
			return "", err
		}
		items := make([]string, l)
		for i := 0; i < l; i++ {
			s, err := decodeAndFormat(sub, elem)
			if err != nil {
				return "", err
			}
			items[i] = s
		}
		return "[" + strings.Join(items, ", ") + "]", nil

	case t.IsFixedSlice():
		elem, l := t.UnSlice()
		var sub *abi.Decoder
		if elem.IsDynamic() {
			var err error
			sub, err = dec.Dynamic()
			if err != nil {
				return "", err
			}
		} else {
			sub = dec
		}
		items := make([]string, l)
		for i := 0; i < l; i++ {
			s, err := decodeAndFormat(sub, elem)
			if err != nil {
				return "", err
			}
			items[i] = s
		}
		return "[" + strings.Join(items, ", ") + "]", nil

	case t.IsTuple():
		args := t.TupleArgs()
		var sub *abi.Decoder
		if t.IsDynamic() {
			var err error
			sub, err = dec.Dynamic()
			if err != nil {
				return "", err
			}
		} else {
			sub = dec
		}
		items := make([]string, len(args))
		for i, a := range args {
			s, err := decodeAndFormat(sub, a)
			if err != nil {
				return "", err
			}
			items[i] = s
		}
		return "(" + strings.Join(items, ", ") + ")", nil

	// --- scalar types ---
	case t == abi.BOOL:
		v, err := dec.Bool()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", v), nil

	case t == abi.ADDRESS:
		v, err := dec.Address()
		if err != nil {
			return "", err
		}
		return v.Hex(), nil

	case t == abi.STRING:
		v, err := dec.DString()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%q", v), nil

	case t == abi.BYTES:
		v, err := dec.Bytes()
		if err != nil {
			return "", err
		}
		return "0x" + hex.EncodeToString(v), nil

	case strings.HasPrefix(st, "bytes"):
		nStr := strings.TrimPrefix(st, "bytes")
		var n int
		fmt.Sscanf(nStr, "%d", &n)
		v, err := dec.ReadNPadRight32(n)
		if err != nil {
			return "", err
		}
		return "0x" + hex.EncodeToString(v), nil

	case strings.HasPrefix(st, "uint"):
		v, err := dec.Uint256()
		if err != nil {
			return "", err
		}
		return v.ToBig().String(), nil

	case strings.HasPrefix(st, "int"):
		v, err := dec.BigInt()
		if err != nil {
			return "", err
		}
		return v.String(), nil

	default:
		return "", fmt.Errorf("unsupported type: %s", t)
	}
}
