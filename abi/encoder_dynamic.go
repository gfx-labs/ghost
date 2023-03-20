package abi

func (d *Builder) String(s string) *Builder {
	dy := d.EnterGroup(len(s), true)
	i := (len(s) / lnlen) * lnlen
	if i > 0 {
		dy.Word([]byte(s[:i]))
	}
	rem := len(s) - i
	if rem > 0 {
		dy.PadRight([]byte(s[i:]))
	}
	return dy.Exit()
}

func (d *Builder) Bytes(s string) *Builder {
	return d.String(s)
}

// fixed size array
// TODO: write in a type + size compliance check later
func (d *Builder) EnterArray(t TypeName, l int) *Builder {
	if t.IsDynamic() {
		return d.EnterGroup(l, false)
	}
	return d.EnterGroup(-l, false)
}

// l = 0 is variable length dynamic
// l = -1 is tuple (static) * # doesnt matter as long as negative
// l > 0 is length specified array (dynamic elements)
// l < 0 is length specified array (static elements)
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

// length unspecified array
func (d *Builder) EnterDynamicArray() *Builder {
	return d.EnterGroup(0, true)
}

func (d *Builder) EnterTuple() *Builder {
	return d.EnterGroup(-1, false)
}

// exit dynamic element
func (d *Builder) Exit() *Builder {
	if d.parent == nil {
		panic("tried to exit group when not in one")
	}
	return d.parent
}
