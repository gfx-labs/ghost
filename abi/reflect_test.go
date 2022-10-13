package abi

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type call_balanceOf struct {
	// input
	Target common.Address `abi:"address,arg"`

	// output
	Balance uint256.Int `abi:"balance"`
}

func TestBalanceOf(t *testing.T) {

}
