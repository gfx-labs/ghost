package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/urfave/cli/v3"
)

var balanceCmd = &cli.Command{
	Name:      "balance",
	Usage:     "fetch the ETH balance of an address",
	ArgsUsage: `<address>`,
	Description: `Fetch the native token (ETH) balance of an address.

Examples:
  ghost balance 0xdead...beef --rpc $RPC
  ghost balance 0xdead...beef --rpc $RPC --block 12345678`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "block",
			Value: "latest",
			Usage: "block number or tag (latest, pending, earliest)",
		},
		&cli.BoolFlag{
			Name:  "wei",
			Usage: "display balance in wei (default: ether)",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() < 1 {
			return fmt.Errorf("usage: ghost balance <address>")
		}

		rpcURL, err := getRPC(cmd)
		if err != nil {
			return err
		}

		addr := cmd.Args().First()
		if !strings.HasPrefix(addr, "0x") {
			addr = "0x" + addr
		}

		result, err := rpcCall(rpcURL, "eth_getBalance", addr, cmd.String("block"))
		if err != nil {
			return err
		}

		var balHex string
		if err := json.Unmarshal(result, &balHex); err != nil {
			return fmt.Errorf("parsing balance: %w", err)
		}

		bal, ok := new(big.Int).SetString(strings.TrimPrefix(balHex, "0x"), 16)
		if !ok {
			return fmt.Errorf("invalid balance hex: %s", balHex)
		}

		if cmd.Bool("wei") {
			fmt.Printf("%s wei\n", bal.String())
		} else {
			// Convert to ether: divide by 1e18
			ether := new(big.Float).Quo(
				new(big.Float).SetInt(bal),
				new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
			)
			fmt.Printf("%s ETH\n", ether.Text('f', 18))
		}
		return nil
	},
}
