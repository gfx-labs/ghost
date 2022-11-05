package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

var (
	// MaxUint256 is the maximum value that can be represented by a uint256.
	MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
	// MaxInt256 is the maximum value that can be represented by a int256.
	MaxInt256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 255), common.Big1)
)

const lnlen = 32 // line length

func pad(data []byte, right bool) []byte {
	l := lnlen * (len(data)/lnlen + 1) // ceiling
	if l == 0 {                        // nil case
		l = lnlen
	}
	padding := make([]byte, l-len(data))
	if right {
		return append(data, padding...)
	} else {
		return append(padding, data...)
	}
}

type Builder struct {
	parent *Builder
	loc    int     // where this starts in the global scope of the encoding
	len    int     // length of segment in bytes
	m      *memory // the encoding of the segment
}

type memory struct {
	encoded []byte // already encoded. history
	cur     int    // current pointer (bytes)
}

func (m *memory) WriteStatic(data []byte) {
	// growing the slice
	new := make([]byte, len(m.encoded)+len(data))
	copy(new, m.encoded[:m.cur])
	copy(new[m.cur+1:], data)
	copy(new[m.cur+len(data):], m.encoded[m.cur:])
	m.encoded = new
	m.cur += len(data)
}

func (m *memory) WriteDynamic(data []byte) {
	m.encoded = append(m.encoded, data...)
}

func (d *Builder) WriteWord(xs []byte) {
	d.m.WriteStatic(pad(xs, false))
	return
}

func (d *Builder) WriteNPadRight32(xs []byte) *Builder {
	d.m.WriteStatic(pad(xs, true))
	return d
}

func (d *Builder) Finish() []byte {
	return d.m.encoded
}

// *************************	WRITING SPECIFIC DATA TYPES
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

// l = length of dynamic element
func (d *Builder) EnterDynamic(l int) *Builder {
	b := &Builder{
		parent: d,
		len:    l,
	}
	b.WriteInt(l)
	return b
}

func (d *Builder) ExitDynamic() *Builder {
	if d.parent == nil {
		panic("tried to exit dynamic when not in one")
	}

	d.parent.m.WriteDynamic(d.m.encoded)
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
