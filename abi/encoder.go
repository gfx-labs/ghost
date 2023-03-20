package abi

import (
	"sort"

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
func (d *Builder) Finish(prefix ...[]byte) []byte {
	if d.children != nil {
		d.children = reorder(d.children)
		for _, c := range d.children {
			c.Finish()
		}
	}
	// if we are a child, there is special logic in our finish
	if d.parent != nil {
		return d.finishChild()
	}
	for _, v := range prefix {
		d.Mem().Insert(0, v)
	}
	return d.Mem().Data()
}

func (d *Builder) finishChild() []byte {
	if d.len == 0 { // length unknown at start
		d.len = len(d.Mem().Data())/lnlen + len(d.children) + d.rlen
	}
	if d.len >= 0 { // dynamic element, need to write offset
		xs := uint256.NewInt(uint64(d.parent.Mem().Pos(0))).Bytes32()
		d.parent.Mem().Put(d.loc, xs[:])
		if d.write {
			d.parent.Int(d.len) // how many elements in the dynamic
			d.parent.rlen -= 1
		}
	}
	d.parent.Mem().Put(-1, d.Mem().Data())
	d.parent.rlen = d.parent.rlen - len(d.Mem().Data())/lnlen
	return d.Mem().Data()
}

func reorder(children []*Builder) []*Builder {
	sort.SliceStable(children, func(i, j int) bool {
		if children[i].len < 0 && children[j].len < 0 {
			return i < j
		}
		if children[i].len < 0 && children[j].len >= 0 {
			return true
		}
		if children[i].len >= 0 && children[j].len < 0 {
			return false
		}
		return i < j
	})
	return children
}

// generic builder writer methods
func (d *Builder) PadRight(xs []byte) *Builder {
	d.Mem().Put(-1, padright(xs))
	return d
}

func (d *Builder) Word(xs []byte) *Builder {
	d.Mem().Put(-1, padleft(xs))
	return d
}
