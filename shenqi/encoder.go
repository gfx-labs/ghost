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
	l := lnlen * (len(data)/(lnlen+1) + 1) // ceiling

	padding := make([]byte, l-len(data))
	if right {
		return append(data, padding...)
	} else {
		return append(padding, data...)
	}
}

type Builder struct {
	parent   *Builder
	len      int    // # of elements for dynamic
	loc      int    // starting pt in the parent builder
	m        memory // the encoding of the segment
	children []*Builder
}

type memory struct {
	encoded []byte // already encoded. history
	cur     int    // current pointer (bytes)
}

func (m *memory) WriteStatic(loc int, data []byte) {
	var s []byte
	if loc == 0 {
		s = data
	} else {
		s = append(m.encoded[:loc], data...)
	}
	if loc == len(m.encoded) {
		m.encoded = s
	} else {
		m.encoded = append(s, m.encoded[loc:]...)
	}
	m.cur += len(data)
}

func (m *memory) WriteDynamic(data []byte) {
	m.encoded = append(m.encoded, data...)
	m.cur += len(data)
}

func (d *Builder) WriteWord(xs []byte) {
	d.m.WriteStatic(d.m.cur, pad(xs, false))
	return
}

// wrinting a dynamic segment's byte location/offset
func (d *Builder) WriteLoc(loc int, i int) {
	xs := big.NewInt(int64(i)).Bytes()
	copy(d.m.encoded[loc:loc+lnlen], pad(xs, false))
}

func (d *Builder) WritePadRight(xs []byte) *Builder {
	d.m.WriteStatic(d.m.cur, pad(xs, true))
	return d
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

func (d *Builder) EnterDynamic(l int) *Builder {
	c := &Builder{
		parent: d,
		loc:    d.m.cur,
		len:    l,
	}
	d.children = append(d.children, c)
	//d.WriteInt(0)
	d.m.encoded = append(d.m.encoded, make([]byte, lnlen)...) // placeholder for line location
	d.m.cur += lnlen
	return c
}

func (d *Builder) ExitDynamic() *Builder {
	if d.parent == nil {
		panic("tried to exit dynamic when not in one")
	}
	return d.parent
}

func (d *Builder) WriteString(s string) *Builder {
	dy := d.EnterDynamic(len(s))
	cur := 0
	for cur+lnlen < len(s) {
		dy.WriteWord([]byte(s[cur:(cur + 32)]))
		cur += lnlen
	}
	if len(s[cur:]) > 0 {
		dy.WritePadRight([]byte(s[cur:]))
	}
	return dy.ExitDynamic()
}

func (d *Builder) writeChild() {
	if d.children != nil {
		for _, c := range d.children {
			c.writeChild()
		}
	}

	if d.parent == nil {
		return
	}

	d.parent.WriteLoc(d.loc, d.parent.m.cur)
	d.parent.WriteInt(d.len)
	d.parent.m.WriteDynamic(d.m.encoded)
}

func (d *Builder) Finish() []byte {
	d.writeChild()
	return d.m.encoded
}
