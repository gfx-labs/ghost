package cli

import (
	"fmt"
	"math/big"
	"strings"

	"gfx.cafe/open/ghost"
	"gfx.cafe/open/ghost/abi"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/urfave/cli/v2"
)

func NewContract(g ghost.Client, name string, addr common.Address) *ContractInteractor {
	return &ContractInteractor{
		c: g, name: name, addr: addr,
	}
}

type ContractInteractor struct {
	name string
	addr common.Address
	c    ghost.Client
}

func (c ContractInteractor) Run(params []string) error {
	app := &cli.App{
		Name:  c.addr.Hex(),
		Usage: "contract interactor",
		Commands: []*cli.Command{
			{
				Name:      "call",
				Aliases:   []string{"c"},
				Usage:     "contract eth_call",
				ArgsUsage: "[abi with parameters]",
				Action: func(ctx *cli.Context) error {
					if ctx.Args().Len() == 0 {
						cli.ShowCommandHelp(ctx, "call")
						return nil
					}
					sig, str := abi.Call(ctx.Args().First()).Decode()
					return c.RunCall(sig, str...)
				},
			},
			{
				Name:      "callargs",
				Aliases:   []string{"ca"},
				Usage:     "contract eth_call",
				ArgsUsage: "[signature] [arguments...]",
				Action: func(ctx *cli.Context) error {
					if ctx.Args().Len() == 0 {
						cli.ShowCommandHelp(ctx, "callargs")
						return nil
					}
					slice := ctx.Args().Slice()
					return c.RunCall(abi.Signature(ctx.Args().First()), slice[1:]...)
				},
			},
		},
	}
	return app.Run(append([]string{"contract"}, params...))
}

func (c ContractInteractor) RunCall(sig abi.Signature, params ...string) error {
	sb := new(strings.Builder)
	for _, v := range sig.Args().TupleArgs() {
		sb.WriteByte('[')
		sb.WriteString(string(v))
		sb.WriteByte(']')
		sb.WriteByte(' ')
	}
	app := &cli.App{
		Name:      sig.Method(),
		Usage:     c.addr.Hex(),
		UsageText: string(sig.Method()) + " " + sb.String(),
		Action: func(ctx *cli.Context) error {
			argz := ctx.Args()
			params := sig.Args().TupleArgs()
			if len(params) > argz.Len() {
				return fmt.Errorf("Too many arguments, Wanted %d got %d", len(params), argz.Len())
			}
			if len(params) < argz.Len() {
				return fmt.Errorf("Too many arguments, Wanted %d got %d", len(params), argz.Len())
			}
			b := new(abi.Builder)
			for idx, typ := range params {
				p := argz.Get(idx)
				switch {
				case typ.IsNumber():
					if typ.IsUnsigned() {
						i, ok := new(big.Int).SetString(p, 0)
						if !ok {
							return fmt.Errorf("arg %d: could not unmarshal %s into number", idx, p)
						}
						b.WriteBigInt(i)
						bi := new(uint256.Int)
						bi.SetFromBig(i)
						b.WriteBigUint(bi)
					} else {
						i, ok := new(big.Int).SetString(p, 0)
						if !ok {
							return fmt.Errorf("arg %d: could not unmarshal %s into number", idx, p)
						}
						b.WriteBigInt(i)
					}
				case typ.IsSlice():
					return fmt.Errorf("slice arguments currently not supported")
				case typ.IsTuple():
					return fmt.Errorf("tuple arguments currently not supported")
				case typ.IsSimple():
					return fmt.Errorf("arg %d: %s arguments currently not supported", idx, typ)
				default:
					return fmt.Errorf("arg %d: unknown abi type %s", idx, typ)
				}
			}
			ans, err := c.c.CallContract(ctx.Context, ethereum.CallMsg{
				To:   &c.addr,
				Data: append(sig.SelectorB(), b.Finish()...),
			}, nil)
			if err != nil {
				return err
			}
			fmt.Printf("%x\n", ans)
			return nil
		},
	}
	return app.Run(append([]string{"eth_call"}, params...))
}
