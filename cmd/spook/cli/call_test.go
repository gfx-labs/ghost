package cli

import (
	"log"
	"testing"

	"gfx.cafe/open/ghost/clients/copypasta"
	"github.com/ethereum/go-ethereum/common"
)

func TestCall(t *testing.T) {
	g, _ := copypasta.Dial("https://mainnet.rpc.gfx.xyz")
	ci := NewContract(g, "ip.vaultcontroller", common.HexToAddress("0x4aae9823fb4c70490f1d802fc697f3fff8d5cbe3"))
	err := ci.RunCall("_vaultId_vaultAddress(uint96)", "2")
	if err != nil {
		log.Println(err)
	}
}
