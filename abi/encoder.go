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
	// insert bytes to a location in memory, appending if needed
	Insert(loc int, data []byte)
	// replacing bytes at a location in memory, growing the slice if needed
	Put(loc int, data []byte)
}

type memory struct {
	encoded []byte // already encoded. history
	cur     int    // current pointer (bytes)
}

func (m *memory) Data() []byte {
	return m.encoded[:m.cur]
}
func (m *memory) Pos(i int) int {
	m.cur = m.cur + i
	return m.cur
}

// inserts data at the location
func (m *memory) Insert(loc int, data []byte) {
	if loc == -1 {
		loc = m.cur
	}
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
	m.Pos(len(data))
}

// replaces data at the location
func (m *memory) Put(loc int, data []byte) {
	if loc == -1 {
		loc = m.cur
	}
	if loc+len(data) > len(m.encoded) {
		m.grow(loc + len(data) - len(m.encoded))
	}
	copy(m.encoded[loc:loc+len(data)], data)
}
func (m *memory) grow(amt int) {
	m.encoded = append(m.encoded, make([]byte, amt)...)
	m.Pos(amt)
}

// *************************	BUILDER
type Builder struct {
	NewMem   func() Memory
	parent   *Builder
	len      int    // # of elements. also used as a boolean
	loc      int    // starting pt in the parent builder
	mm       Memory // the encoding of the segment
	bm       memory
	children []*Builder
	rlen     int  // running length
	write    bool // write length or not
}

// get the memory object, uses default memory impl by default
func (d *Builder) Mem() Memory {
	if d.mm != nil {
		if d.NewMem != nil {
			d.mm = d.NewMem()
		}
		return d.mm
	}
	return &d.bm
}

// l = 0 is variable length dynamic
// l = -1 is tuple (static)
// l > 0 is length specified array (dynamic elements)
// l < 0 is length specified array (static elements)
func (d *Builder) EnterGroup(l int, w bool) *Builder {
	c := &Builder{
		parent: d,
		loc:    d.Mem().Pos(0),
		len:    l,
		NewMem: d.NewMem,
		write:  w, // whether to write length
	}
	d.children = append(d.children, c)
	if l >= 0 { // is dynamic
		wd := [32]byte{} // insert offset placeholder
		d.Mem().Put(-1, wd[:])
		d.rlen -= 1
		if d.len == -1 { // parent is static tuple
			d.len = 0
			if d.parent != nil {
				d.parent.Mem().Insert(d.loc, wd[:])
				d.parent.rlen = d.parent.rlen - 1
			}
			b := d.parent
			for b.parent != nil && b.parent.len == -1 {
				b.parent.Mem().Insert(b.loc, wd[:])
				b.parent.rlen = b.parent.rlen - 1
				b = b.parent
			}
		}
	}
	return c
}

// length unspecified array
func (d *Builder) EnterDynamicArray() *Builder {
	return d.EnterGroup(0, true)
}

func (d *Builder) EnterTuple() *Builder {
	return d.EnterGroup(-1, false)
}

// fixed size array
// TODO: write in a type + size compliance check later
func (d *Builder) EnterArray(t TypeName, l int) *Builder {
	if t.IsDynamic() {
		return d.EnterGroup(l, false)
	}
	return d.EnterGroup(-l, false)
}

// exit dynamic element
func (d *Builder) Exit() *Builder {
	if d.parent == nil {
		panic("tried to exit group when not in one")
	}
	return d.parent
}

func reorder(children []*Builder) []*Builder {
	r := make([]*Builder, len(children))
	border := 0
	for _, c := range children {
		if c.len < 0 { // is static
			border++
		}
	}
	i := 0
	j := 0
	for _, c := range children {
		if c.len < 0 {
			r[i] = c
			i++
		} else {
			r[border+j] = c
			j++
		}
	}
	return r
}

// finish closes all children, and returns the result slice
func (d *Builder) Finish() []byte {
	if d.children != nil {
		d.children = reorder(d.children)
		for _, c := range d.children {
			c.Finish()
		}
	}
	if d.parent == nil {
		return d.Mem().Data()
	}
	if d.len == 0 {
		d.len = len(d.Mem().Data())/lnlen + len(d.children) + d.rlen
	}
	if d.len >= 0 { // dynamic element, need to write offset
		xs := uint256.NewInt(uint64(d.parent.Mem().Pos(0))).Bytes32()
		d.parent.Mem().Put(d.loc, xs[:])
		//d.parent.rlen -= 1
		if d.write {
			d.parent.WriteInt(d.len) // how many elements in the dynamic
			d.parent.rlen -= 1
		}
	}
	d.parent.Mem().Put(-1, d.Mem().Data())
	d.parent.rlen = d.parent.rlen - len(d.Mem().Data())/lnlen
	return d.Mem().Data()
}

// generic builder writer methods
func (d *Builder) WritePadRight(xs []byte) *Builder {
	d.Mem().Put(-1, padright(xs))
	return d
}

func (d *Builder) WriteWord(xs []byte) *Builder {
	d.Mem().Put(-1, padleft(xs))
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
func (d *Builder) WriteFixedBytes(i int, s string) *Builder {
	if i < len(s) {
		panic("input length mismatch")
	}
	return d.WritePadRight([]byte(s))
}

func (d *Builder) WriteString(s string) *Builder {
	dy := d.EnterGroup(len(s), true)
	i := (len(s) / lnlen) * lnlen
	if i > 0 {
		dy.WriteWord([]byte(s[:i]))
	}
	rem := len(s) - i
	if rem > 0 {
		dy.WritePadRight([]byte(s[i:]))
	}
	return dy.Exit()
}

func (d *Builder) WriteBytes(s string) *Builder {
	return d.WriteString(s)
}
