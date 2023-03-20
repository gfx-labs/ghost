package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func (d *Builder) WriteBigUint(a *uint256.Int) *Builder {
	d.WriteWord(a.Bytes())
	return d
}

func (d *Builder) WriteAddress(a common.Address) *Builder {
	d.WriteWord(a[:])
	return d
}

func (d *Builder) WriteBigInt(ret *big.Int) *Builder {
	nr := new(big.Int).Set(ret)
	if ret.Cmp(new(big.Int)) == -1 {
		nr.Neg(nr)
		nr.Sub(nr, common.Big1)
		nr.Sub(nr, MaxUint256)
		nr.Neg(nr)
		d.WriteWord(nr.Bytes())
		return d
	}
	d.WriteWord(ret.Bytes())
	return d
}

func (d *Builder) WriteBool(b bool) *Builder {
	if b == true {
		d.WriteWord([]byte{0x1})
		return d
	}
	d.WriteWord([]byte{0x0})
	return d
}

func (d *Builder) WriteInt(i int) *Builder {
	if i >= 0 {
		return d.WriteUint(uint(i))
	}
	return d.WriteBigInt(big.NewInt(int64(i)))
}

func (d *Builder) WriteUint(i uint) *Builder {
	return d.WriteBigUint(uint256.NewInt(uint64(i)))
}

func (d *Builder) WriteUint8(i uint8) *Builder {
	return d.WriteUint(uint(i))
}

func (d *Builder) WriteUint16(i uint16) *Builder {
	d.WriteUint(uint(i))
	return d
}

// 0 < i <= 32
func (d *Builder) WriteFixedBytes(l int, s []byte) *Builder {
	//fmt.Printf("%v %s %v\n", l, s, len(s))
	if l < len(s) {
		panic("input length mismatch")
	}
	return d.WritePadRight(s)
}
