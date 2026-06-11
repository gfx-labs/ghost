package abi

import (
	"sort"

	"github.com/holiman/uint256"
)

const lnlen = 32 // line length

// Builder constructs ABI-encoded byte sequences using a fluent API.
//
// Static values (uint, address, bool, fixed bytes) are written inline with
// methods like [Builder.Uint], [Builder.Address], and [Builder.Bool].
//
// Dynamic values (strings, byte slices, dynamic arrays, tuples) require
// entering a group with [Builder.EnterDynamicArray], [Builder.EnterTuple],
// or [Builder.EnterArray], writing the contents, then calling [Builder.Exit]
// to return to the parent context. The Builder handles offset calculation
// and length prefixing automatically.
//
// Call [Builder.Finish] on the root builder to produce the final encoded bytes.
// An optional prefix (typically a 4-byte function selector from [Signature.Fn])
// can be prepended.
//
// A zero-value Builder is ready to use:
//
//	b := new(abi.Builder)
//	b.Uint(42).DString("hello")
//	encoded := b.Finish()
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

// AbiBuilderOpt is a functional option for configuring a [Builder].
type AbiBuilderOpt func(*Builder) *Builder

// WithBuilderMemory returns an option that sets a custom [Memory] factory
// for the builder. By default, Builder uses an internal slice-based memory.
func WithBuilderMemory(fn func() Memory) AbiBuilderOpt {
	return func(d *Builder) *Builder {
		d.NewMem = fn
		return d
	}
}

// NewBuilder creates a Builder with the given options.
// A zero-value Builder (&Builder{}) is also valid.
func NewBuilder(opts ...AbiBuilderOpt) *Builder {
	b := &Builder{}
	for _, v := range opts {
		if v != nil {
			b = v(b)
		}
	}
	return b
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

// PadRight writes xs zero-padded on the right to 32 bytes.
// Used for bytesN and other right-aligned types.
func (d *Builder) PadRight(xs []byte) *Builder {
	d.Mem().Put(-1, padright(xs))
	return d
}

// Word writes xs zero-padded on the left to 32 bytes.
// Used for uint, int, address, and other left-aligned types.
func (d *Builder) Word(xs []byte) *Builder {
	d.Mem().Put(-1, padleft(xs))
	return d
}

// Finish finalizes the encoding, resolves all dynamic offsets, and returns
// the encoded bytes. Optional prefix slices (typically a 4-byte function
// selector) are prepended to the output.
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
