package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type Encoder struct {
	xs  []byte
	cur int
}

func NewEncoder(xs []byte) *Encoder {
	return &Encoder{
		xs: xs,
	}
}

func (d *Encoder) Write(o []byte) (int, error) {
	d.xs = append(d.xs, o...)
	return len(o), nil
}

func (d *Encoder) WriteNPadRight32(xs []byte) error {
	diff := 32 - (len(xs) % 32)
	if diff == 32 {
		diff = 0
	}
	o := append(xs, make([]byte, diff)...)
	_, err := d.Write(o[:])
	if err != nil {
		return err
	}
	return nil
}
func (d *Encoder) WriteWord(xs []byte) error {
	diff := 32 - (len(xs) % 32)
	if diff == 32 {
		diff = 0
	}
	o := append(make([]byte, diff), xs...)
	_, err := d.Write(o[:])
	if err != nil {
		return err
	}
	return nil
}

func (d *Encoder) WriteBigUint(a *uint256.Int) error {
	return d.WriteWord(a.Bytes())
}

func (d *Encoder) WriteAddress(a common.Address) error {
	return d.WriteWord(a[:])
}

func (d *Encoder) WriteBigInt(ret *big.Int) error {
	nr := new(big.Int).Set(ret)
	if ret.Cmp(new(big.Int)) == -1 {
		nr.Neg(nr)
		nr.Sub(nr, common.Big1)
		nr.Sub(nr, MaxUint256)
		nr.Neg(nr)
		return d.WriteWord(nr.Bytes())
	}
	return d.WriteWord(ret.Bytes())
}

func (d *Encoder) WriteBool(b bool) error {
	if b == true {
		return d.WriteWord([]byte{0x1})
	}
	return d.WriteWord([]byte{0x0})
}

func (d *Encoder) WriteInt(i int) error {
	return d.WriteBigInt(big.NewInt(int64(i)))
}

func (d *Encoder) WriteUint(i uint) error {
	return d.WriteBigUint(uint256.NewInt(uint64(i)))
}

func (d *Encoder) WriteUint8(i uint8) error {
	return d.WriteUint(uint(i))
}

func (d *Encoder) WriteUint16(i uint16) error {
	return d.WriteUint(uint(i))
}

func (d *Encoder) Finish() []byte {
	return d.xs
}
func (d *Encoder) Reset() {
	d.xs = d.xs[:0]
}

func (d *Encoder) WriteDynamic(*Encoder) error {
	panic("not done")
}

func (d *Encoder) WriteString() (string, error) {
	panic("not done")
}
