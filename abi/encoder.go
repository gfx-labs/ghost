package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type Builder struct {
	parent *Builder

	cur  int
	args [][2][]byte

	dsz int

	prefix []byte
}

func NewBuilder(xs []byte) *Builder {
	return &Builder{prefix: xs}
}

func (d *Builder) WriteNPadRight32(xs []byte) *Builder {
	diff := 32 - (len(xs) % 32)
	if diff == 32 {
		diff = 0
	}
	o := append(xs, make([]byte, diff)...)
	d.args = append(d.args, [2][]byte{o, nil})
	return d
}
func (d *Builder) WriteWord(xs []byte) {
	diff := 32 - (len(xs) % 32)
	if diff == 32 {
		diff = 0
	}
	o := append(make([]byte, diff), xs...)
	d.args = append(d.args, [2][]byte{o, nil})
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
	out := make([]byte, 0, 32*len(d.args)+d.dsz)
	out = append(out, d.prefix...)
	free := len(d.args) * 32
	for _, v := range d.args {
		if len(v[0]) == 0 {
			b32 := uint256.NewInt(uint64(free)).Bytes32()
			out = append(out, b32[:]...)
		} else {
			out = append(out, v[0]...)
		}
		free = free + len(v[1])
	}
	for _, v := range d.args {
		out = append(out, v[1]...)
	}
	d.args = d.args[:0]
	return out
}

func (d *Builder) Reset() {
	d.args = d.args[:0]
}

func (d *Builder) EnterDynamic(l int) *Builder {
	b := &Builder{parent: d}
	if l > 0 {
		b32 := uint256.NewInt(uint64(l)).Bytes32()
		b.prefix = b32[:]
	}
	return b
}
func (d *Builder) ExitDynamic() *Builder {
	if d.parent == nil {
		panic("tried to exit dynamic when not in one")
	}
	bts := d.Finish()
	d.parent.args = append(d.parent.args, [2][]byte{nil, bts})
	d.dsz = d.dsz + len(bts)
	return d.parent
}

func (d *Builder) WriteString(s string) *Builder {
	dy := d.EnterDynamic(len(s))
	cur := 0
	for {
		if cur+32 >= len(s) {
			if len(s[cur:]) > 0 {
				dy.WriteNPadRight32([]byte(s[cur:]))
			}
			break
		} else {
			dy.WriteWord([]byte(s[cur:(cur + 32)]))
			cur = cur + 32
		}
	}
	return dy.ExitDynamic()
}
