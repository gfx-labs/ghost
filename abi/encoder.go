package abi

import (
	"github.com/holiman/uint256"
)

const lnlen = 32 // line length

// Builder is used to encode ethereum abi
type Builder struct {
	NewMem   func() Memory
	parent   *Builder
	len      int    // # of elements. 0 value is special case
	loc      int    // starting pt in the parent builder
	mm       Memory // the encoding of the segment
	bm       sliceMemory
	children []*Builder
	rlen     int  // running length. to not double count dynamic children
	write    bool // whether to write length
}

func NewBuilder(fn func() Memory) *Builder {
	return &Builder{
		NewMem: fn,
	}
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
	if d.len == 0 { // length unknown at start
		d.len = len(d.Mem().Data())/lnlen + len(d.children) + d.rlen
	}
	if d.len >= 0 { // dynamic element, need to write offset
		xs := uint256.NewInt(uint64(d.parent.Mem().Pos(0))).Bytes32()
		d.parent.Mem().Put(d.loc, xs[:])
		if d.write {
			d.parent.WriteInt(d.len) // how many elements in the dynamic
			d.parent.rlen -= 1
		}
	}
	d.parent.Mem().Put(-1, d.Mem().Data())
	d.parent.rlen = d.parent.rlen - len(d.Mem().Data())/lnlen
	return d.Mem().Data()
}

func reorder(children []*Builder) []*Builder {
	r := make([]*Builder, len(children))
	border := 0
	for _, c := range children {
		if c.len < 0 { // is static
			border++
		}
	}
	var i, j int
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

// generic builder writer methods
func (d *Builder) WritePadRight(xs []byte) *Builder {
	d.Mem().Put(-1, padright(xs))
	return d
}

func (d *Builder) WriteWord(xs []byte) *Builder {
	d.Mem().Put(-1, padleft(xs))
	return d
}
