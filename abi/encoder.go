package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type Builder struct {
	xs  []byte
	cur int
}

func NewBuilder(xs []byte) *Builder {
	return &Builder{
		xs: xs,
	}
}

func (d *Builder) Write(o []byte) *Builder {
	d.xs = append(d.xs, o...)
	return d
}

func (d *Builder) WriteNPadRight32(xs []byte) *Builder {
	diff := 32 - (len(xs) % 32)
	if diff == 32 {
		diff = 0
	}
	o := append(xs, make([]byte, diff)...)
	d.Write(o[:])
	return d
}
func (d *Builder) WriteWord(xs []byte) {
	diff := 32 - (len(xs) % 32)
	if diff == 32 {
		diff = 0
	}
	o := append(make([]byte, diff), xs...)
	d.Write(o[:])
	return
}

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
	d.WriteBigInt(big.NewInt(int64(i)))
	return d
}

func (d *Builder) WriteUint(i uint) *Builder {
	d.WriteBigUint(uint256.NewInt(uint64(i)))
	return d
}

func (d *Builder) WriteUint8(i uint8) *Builder {
	d.WriteUint(uint(i))
	return d
}

func (d *Builder) WriteUint16(i uint16) *Builder {
	d.WriteUint(uint(i))
	return d
}

func (d *Builder) Finish() []byte {
	return d.xs
}
func (d *Builder) Reset() {
	d.xs = d.xs[:0]
}

func (d *Builder) StartDynamic(*Builder) *Builder {
	panic("not done")
}
func (d *Builder) FinishDynamic(*Builder) *Builder {
	panic("not done")
}

func (d *Builder) WriteString(string) *Builder {
	panic("not done")
}
