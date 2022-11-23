package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

const lnlen = 32 // line length
type Memory interface {
	// returns the full data
	Data() []byte
	// increments the cursor by input, returns new cursor
	Pos(int) int

	// write static bytes to memoryu
	WriteStatic(data []byte)
	// write bytes to a location in memory
	WriteLoc(loc int, i int)
	// write to the heap
	WriteHeap(data []byte)
}

type Builder struct {
	NewMem func() Memory
	parent *Builder
	len    int // # of elements for dynamic
	loc    int // starting pt in the parent builder

	mm       Memory // the encoding of the segment
	bm       memory
	children []*Builder
}

// default memory implementation
type memory struct {
	encoded []byte // already encoded. history
	cur     int    // current pointer (bytes)
}

func (m *memory) WriteStatic(data []byte) {
	var s []byte
	if m.cur == 0 {
		s = data
	} else {
		s = append(m.encoded[:m.cur], data...)
	}
	if len(m.encoded) == m.cur {
		m.encoded = s
	} else {
		m.encoded = append(s, m.encoded[m.cur:]...)
	}
	m.Pos(len(data))
}

func (m *memory) WriteHeap(data []byte) {
	m.encoded = append(m.encoded, data...)
	m.Pos(len(data))
}

func (m *memory) Data() []byte {
	return m.encoded
}
func (m *memory) Pos(i int) int {
	m.cur = m.cur + i
	return m.cur
}

// wrinting a dynamic segment's byte location/offset
func (m *memory) WriteLoc(loc int, i int) {
	xs := uint256.NewInt(uint64(i)).Bytes()
	copy(m.encoded[loc:loc+lnlen], padleft(xs))
}

// get the memory object
func (d *Builder) Mem() Memory {
	if d.mm != nil {
		if d.NewMem != nil {
			d.mm = d.NewMem()
		}
		return d.mm
	}
	return &d.bm
}

// generic builder writer methods
func (d *Builder) WritePadRight(xs []byte) *Builder {
	d.Mem().WriteStatic(padright(xs))
	return d
}

func (d *Builder) WriteWord(xs []byte) *Builder {
	d.Mem().WriteStatic(padleft(xs))
	return d
}

// builder dynamic handling
func (d *Builder) EnterDynamic(l int) *Builder {
	c := &Builder{
		parent: d,
		loc:    d.Mem().Pos(0),
		len:    l,
		NewMem: d.NewMem,
	}
	d.children = append(d.children, c)
	if l > 0 {
		wd := [32]byte{}
		d.Mem().WriteStatic(wd[:])
	}
	return c
}

func (d *Builder) Dynamic() *Builder {
	return d.EnterDynamic(-1)
}

func (d *Builder) ExitDynamic() *Builder {
	if d.parent == nil {
		panic("tried to exit dynamic when not in one")
	}
	return d.parent
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

func (d *Builder) Finish() []byte {
	if d.children != nil {
		for _, c := range d.children {
			c.Finish()
		}
	}
	if d.parent == nil {
		return d.Mem().Data()
	}
	d.parent.Mem().WriteLoc(d.loc, d.parent.Mem().Pos(0))
	if d.len > 0 {
		if d.len < 1 {
			d.len = len(d.children)
		}
		d.parent.WriteInt(d.len)
	}
	d.parent.Mem().WriteHeap(d.Mem().Data())
	return d.Mem().Data()
}

//func (d *Builder) Finish() []byte {
//	d.writeChild()
//	return d.Mem().Data()
//}
