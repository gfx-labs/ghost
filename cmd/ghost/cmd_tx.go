package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gfx-labs/ghost/abi"
	"github.com/urfave/cli/v3"
)

var txCmd = &cli.Command{
	Name:      "tx",
	Usage:     "fetch a transaction and display its details",
	ArgsUsage: `<hash>`,
	Description: `Fetch a transaction by hash and display its details.
Shows from, to, value, gas, calldata, and optionally decodes calldata.

Examples:
  ghost tx 0xabc... --rpc $RPC
  ghost tx 0xabc... --rpc $RPC --decode "transfer(address,uint256)"`,
	Flags: []cli.Flag{
		outputFlag,
		&cli.StringFlag{
			Name:    "decode",
			Aliases: []string{"d"},
			Usage:   "function signature to decode the calldata",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() < 1 {
			return fmt.Errorf("usage: ghost tx <hash>")
		}

		rpcURL, err := getRPC(cmd)
		if err != nil {
			return err
		}

		hash := cmd.Args().First()
		if !strings.HasPrefix(hash, "0x") {
			hash = "0x" + hash
		}

		result, err := rpcCall(rpcURL, "eth_getTransactionByHash", hash)
		if err != nil {
			return err
		}

		var tx struct {
			Hash             string `json:"hash"`
			From             string `json:"from"`
			To               string `json:"to"`
			Value            string `json:"value"`
			Gas              string `json:"gas"`
			GasPrice         string `json:"gasPrice"`
			Input            string `json:"input"`
			Nonce            string `json:"nonce"`
			BlockNumber      string `json:"blockNumber"`
			BlockHash        string `json:"blockHash"`
			TransactionIndex string `json:"transactionIndex"`
		}
		if err := json.Unmarshal(result, &tx); err != nil {
			return fmt.Errorf("parsing transaction: %w", err)
		}

		fmt.Printf("hash:        %s\n", tx.Hash)
		fmt.Printf("from:        %s\n", tx.From)
		fmt.Printf("to:          %s\n", tx.To)
		fmt.Printf("value:       %s\n", tx.Value)
		fmt.Printf("gas:         %s\n", tx.Gas)
		fmt.Printf("gasPrice:    %s\n", tx.GasPrice)
		fmt.Printf("nonce:       %s\n", tx.Nonce)
		fmt.Printf("block:       %s\n", tx.BlockNumber)

		if tx.Input != "" && tx.Input != "0x" {
			fmt.Printf("input:       %s\n", tx.Input)
			if len(tx.Input) >= 10 {
				fmt.Printf("selector:    %s\n", tx.Input[:10])
			}

			// Decode calldata if signature provided
			if d := cmd.String("decode"); d != "" {
				data, err := parseHex(tx.Input)
				if err != nil {
					return err
				}
				if len(data) < 4 {
					return fmt.Errorf("calldata too short")
				}
				fmt.Printf("selector:    0x%s\n", hex.EncodeToString(data[:4]))
				fmt.Println("--- decoded calldata ---")
				// Reuse decode logic
				typeStr := d
				var typeName string
				if i := strings.Index(typeStr, "("); i > 0 && typeStr[0] != '(' {
					typeName = typeStr
				} else {
					typeName = "fn" + typeStr // make it a function sig
				}
				_ = typeName

				// Parse as tuple from the function signature
				sig := d
				if i := strings.Index(sig, "("); i > 0 {
					tupleStr := sig[strings.Index(sig, "("):]
					types := typeName // use full sig for decoding
					_ = types
					args := (abi.TypeName(tupleStr)).TupleArgs()
					dec := abi.NewDecoder(data[4:])
					for i, t := range args {
						val, err := decodeAndFormat(dec, t)
						if err != nil {
							fmt.Printf("[%d] %s: error: %v\n", i, t, err)
							break
						}
						fmt.Printf("[%d] %s: %s\n", i, t, val)
					}
				}
			}
		}
		return nil
	},
}
