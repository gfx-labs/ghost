package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/urfave/cli/v3"
)

var blockCmd = &cli.Command{
	Name:      "block",
	Usage:     "fetch block information",
	ArgsUsage: `[number|tag]`,
	Description: `Fetch block information by number or tag.

Examples:
  ghost block --rpc $RPC               # latest block
  ghost block latest --rpc $RPC
  ghost block 12345678 --rpc $RPC
  ghost block 0xabc123 --rpc $RPC      # by hex number`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "full",
			Usage: "include full transaction objects (default: hashes only)",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		rpcURL, err := getRPC(cmd)
		if err != nil {
			return err
		}

		blockID := "latest"
		if cmd.NArg() > 0 {
			blockID = cmd.Args().First()
		}

		// Convert decimal to hex if needed
		switch blockID {
		case "latest", "pending", "earliest", "safe", "finalized":
			// tag, use as-is
		default:
			if !strings.HasPrefix(blockID, "0x") {
				n, ok := new(big.Int).SetString(blockID, 10)
				if ok {
					blockID = "0x" + n.Text(16)
				}
			}
		}

		result, err := rpcCall(rpcURL, "eth_getBlockByNumber", blockID, cmd.Bool("full"))
		if err != nil {
			return err
		}

		var block struct {
			Number       string   `json:"number"`
			Hash         string   `json:"hash"`
			ParentHash   string   `json:"parentHash"`
			Timestamp    string   `json:"timestamp"`
			GasUsed      string   `json:"gasUsed"`
			GasLimit     string   `json:"gasLimit"`
			BaseFee      string   `json:"baseFeePerGas"`
			Miner        string   `json:"miner"`
			Transactions []any    `json:"transactions"`
			Size         string   `json:"size"`
		}
		if err := json.Unmarshal(result, &block); err != nil {
			return fmt.Errorf("parsing block: %w", err)
		}

		blockNum, _ := new(big.Int).SetString(strings.TrimPrefix(block.Number, "0x"), 16)
		ts, _ := new(big.Int).SetString(strings.TrimPrefix(block.Timestamp, "0x"), 16)
		gasUsed, _ := new(big.Int).SetString(strings.TrimPrefix(block.GasUsed, "0x"), 16)
		gasLimit, _ := new(big.Int).SetString(strings.TrimPrefix(block.GasLimit, "0x"), 16)

		fmt.Printf("number:       %s\n", blockNum)
		fmt.Printf("hash:         %s\n", block.Hash)
		fmt.Printf("parent:       %s\n", block.ParentHash)
		fmt.Printf("timestamp:    %s\n", ts)
		fmt.Printf("miner:        %s\n", block.Miner)
		fmt.Printf("gasUsed:      %s\n", gasUsed)
		fmt.Printf("gasLimit:     %s\n", gasLimit)
		if block.BaseFee != "" {
			baseFee, _ := new(big.Int).SetString(strings.TrimPrefix(block.BaseFee, "0x"), 16)
			fmt.Printf("baseFee:      %s\n", baseFee)
		}
		fmt.Printf("transactions: %d\n", len(block.Transactions))
		return nil
	},
}
