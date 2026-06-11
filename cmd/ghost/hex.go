package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

// readHexInput reads hex data from: positional arg, -i file, or stdin.
// The argIdx parameter specifies which positional arg to try.
func readHexInput(cmd *cli.Command, argIdx int) ([]byte, error) {
	// Try -i file first
	if f := cmd.String("input"); f != "" {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("reading input file: %w", err)
		}
		return parseHex(string(data))
	}

	// Try positional arg
	if cmd.NArg() > argIdx {
		return parseHex(cmd.Args().Get(argIdx))
	}

	// Try stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		return parseHex(string(data))
	}

	return nil, fmt.Errorf("no hex input provided (pass as argument, pipe to stdin, or use -i)")
}

// writeOutput writes bytes as hex to: -o file or stdout.
func writeOutput(cmd *cli.Command, data []byte) error {
	out := "0x" + hex.EncodeToString(data)
	if f := cmd.String("output"); f != "" {
		return os.WriteFile(f, []byte(out+"\n"), 0644)
	}
	fmt.Println(out)
	return nil
}

// writeString writes a string to: -o file or stdout.
func writeString(cmd *cli.Command, s string) error {
	if f := cmd.String("output"); f != "" {
		return os.WriteFile(f, []byte(s+"\n"), 0644)
	}
	fmt.Println(s)
	return nil
}

// parseHex decodes a hex string, stripping 0x prefix and whitespace.
func parseHex(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	if len(s) == 0 {
		return []byte{}, nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return b, nil
}

// common I/O flags for commands that read/write hex
var inputFlag = &cli.StringFlag{
	Name:    "input",
	Aliases: []string{"i"},
	Usage:   "read hex input from file",
}

var outputFlag = &cli.StringFlag{
	Name:    "output",
	Aliases: []string{"o"},
	Usage:   "write output to file",
}
