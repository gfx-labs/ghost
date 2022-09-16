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

func (d *Builder) Write(o []byte) {
	d.xs = append(d.xs, o...)
}

func (d *Builder) WriteNPadRight32(xs []byte) {
	diff := 32 - (len(xs) % 32)
	if diff == 32 {
		diff = 0
	}
	o := append(xs, make([]byte, diff)...)
	d.Write(o[:])
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

func (d *Builder) WriteBigUint(a *uint256.Int) {
	d.WriteWord(a.Bytes())
}

func (d *Builder) WriteAddress(a common.Address) {
	d.WriteWord(a[:])
}

func (d *Builder) WriteBigInt(ret *big.Int) {
	nr := new(big.Int).Set(ret)
	if ret.Cmp(new(big.Int)) == -1 {
		nr.Neg(nr)
		nr.Sub(nr, common.Big1)
		nr.Sub(nr, MaxUint256)
		nr.Neg(nr)
		d.WriteWord(nr.Bytes())
		return
	}
	d.WriteWord(ret.Bytes())
}

func (d *Builder) WriteBool(b bool) {
	if b == true {
		d.WriteWord([]byte{0x1})
		return
	}
	d.WriteWord([]byte{0x0})
}

func (d *Builder) WriteInt(i int) {
	d.WriteBigInt(big.NewInt(int64(i)))
}

func (d *Builder) WriteUint(i uint) {
	d.WriteBigUint(uint256.NewInt(uint64(i)))
}

func (d *Builder) WriteUint8(i uint8) {
	d.WriteUint(uint(i))
}

func (d *Builder) WriteUint16(i uint16) {
	d.WriteUint(uint(i))
}

func (d *Builder) Finish() []byte {
	return d.xs
}
func (d *Builder) Reset() {
	d.xs = d.xs[:0]
}

func (d *Builder) WriteDynamic(*Builder) {
	panic("not done")
}

func (d *Builder) WriteString(string) {
	panic("not done")
}
