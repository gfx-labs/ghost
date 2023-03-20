package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func (d *Builder) Uint256(a *uint256.Int) *Builder {
	d.Word(a.Bytes())
	return d
}

func (d *Builder) Address(a common.Address) *Builder {
	d.Word(a[:])
	return d
}

func (d *Builder) BigInt(ret *big.Int) *Builder {
	nr := new(big.Int).Set(ret)
	if ret.Cmp(new(big.Int)) == -1 {
		nr.Neg(nr)
		nr.Sub(nr, common.Big1)
		nr.Sub(nr, MaxUint256)
		nr.Neg(nr)
		d.Word(nr.Bytes())
		return d
	}
	d.Word(ret.Bytes())
	return d
}

func (d *Builder) Bool(b bool) *Builder {
	if b == true {
		d.Word([]byte{0x1})
		return d
	}
	d.Word([]byte{0x0})
	return d
}

func (d *Builder) Int(i int) *Builder {
	if i >= 0 {
		return d.Uint(uint(i))
	}
	return d.BigInt(big.NewInt(int64(i)))
}

func (d *Builder) Uint(i uint) *Builder {
	return d.BigUint(uint256.NewInt(uint64(i)))
}

func (d *Builder) Uint8(i uint8) *Builder {
	return d.Uint(uint(i))
}

func (d *Builder) Uint16(i uint16) *Builder {
	d.Uint(uint(i))
	return d
}

// 0 < i <= 32
func (d *Builder) FixedBytes(l int, s []byte) *Builder {
	//fmt.Printf("%v %s %v\n", l, s, len(s))
	if l < len(s) {
		panic("input length mismatch")
	}
	return d.PadRight(s)
}

// DEPRECATED
func (d *Builder) BigUint(a *uint256.Int) *Builder {
	return d.Uint256(a)
}
