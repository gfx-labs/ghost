package main

import (
	"os"

	"gfx.cafe/open/ghost/clients/copypasta"
	"gfx.cafe/open/ghost/cmd/spook/cli"
)

func main() {
	c, _ := copypasta.Dial("https://mainnet.rpc.gfx.xyz")
	app := cli.NewCli(c)
	app.Run(os.Args)
}
