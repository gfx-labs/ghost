package abi

// DString encodes a dynamic string. The ABI representation includes an offset,
// a length prefix, and the string data padded to 32-byte alignment.
func (d *Builder) DString(s string) *Builder {
	return d.Bytes([]byte(s))
}

// Bytes encodes a dynamic byte sequence. Like [Builder.DString] but accepts []byte.
func (d *Builder) Bytes(s []byte) *Builder {
	dy := d.EnterGroup(len(s), true)
	i := (len(s) / lnlen) * lnlen
	if i > 0 {
		dy.Word(s[:i])
	}
	rem := len(s) - i
	if rem > 0 {
		dy.PadRight(s[i:])
	}
	return dy.Exit()
}

// EnterArray begins a fixed-size array of l elements with element type t.
// If t is dynamic, the array is treated as a dynamic group (with offsets);
// otherwise it is treated as a static group (elements inline).
// Write exactly l elements, then call [Builder.Exit].
func (d *Builder) EnterArray(t TypeName, l int) *Builder {
	if t.IsDynamic() {
		return d.EnterGroup(l, false)
	}
	return d.EnterGroup(-l, false)
}

// EnterGroup begins a child encoding group. This is the low-level primitive
// behind [Builder.EnterDynamicArray], [Builder.EnterTuple], and [Builder.EnterArray].
//
// The l parameter controls group behavior:
//   - l == 0: variable-length dynamic (length written on finish)
//   - l > 0:  fixed-length dynamic (l elements, with offset pointer)
//   - l < 0:  static group (|l| elements, no offset pointer — used for tuples and static arrays)
//
// The w parameter controls whether the element count is written as a prefix
// when the group is finalized (true for dynamic arrays and bytes, false for
// tuples and fixed arrays).
func (d *Builder) EnterGroup(l int, w bool) *Builder {
	c := &Builder{
		parent: d,
		loc:    d.Mem().Pos(0),
		len:    l,
		NewMem: d.NewMem,
		write:  w,
	}
	d.children = append(d.children, c)
	if l >= 0 { // is dynamic
		wd := [32]byte{} // insert offset placeholder
		d.Mem().Put(-1, wd[:])
		d.rlen -= 1
		b := d
		for b.len < 0 {
			b.len = -b.len
			if b.parent != nil {
				b.parent.Mem().Insert(b.loc, wd[:])
				b.parent.rlen = b.parent.rlen - 1
				b = b.parent
			}
		}
	}
	return c
}

// EnterDynamicArray begins a variable-length dynamic array (e.g. uint256[]).
// Write any number of elements, then call [Builder.Exit]. The length is
// determined automatically and written as a prefix.
func (d *Builder) EnterDynamicArray() *Builder {
	return d.EnterGroup(0, true)
}

// EnterTuple begins a tuple group. Write each tuple field in order,
// then call [Builder.Exit]. Tuples are encoded as static groups unless
// they contain dynamic elements, in which case the Builder handles
// offset insertion automatically.
func (d *Builder) EnterTuple() *Builder {
	return d.EnterGroup(-1, false)
}

// Exit closes the current group and returns the parent builder.
// Panics if called on a root builder (not inside a group).
func (d *Builder) Exit() *Builder {
	if d.parent == nil {
		panic("tried to exit group when not in one")
	}
	return d.parent
}
