package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type Word = uint256.Int

func AddressToWord(c common.Address) Word {
	w := Word{}
	w.SetBytes(c.Bytes())
	return w
}
