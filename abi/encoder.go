package abi

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// the word, which contains either pointer or bytes data
type word struct {
	pointer int
	dat     []byte
}

func padRight32(xs []byte) []byte {
	pl := 32 * (len(xs) / 32)
	if pl == 0 {
		pl = 32
	}
	return rightPadBytes(xs, pl)
}
func padLeft32(xs []byte) []byte {
	pl := 32 * (len(xs) / 32)
	if pl == 0 {
		pl = 32
	}
	r := leftPadBytes(xs, pl)
	return r
}

func rightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}
	padded := make([]byte, l-len(slice))
	return append(slice, padded...)
}

func leftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}
	padded := make([]byte, l-len(slice))
	return append(padded, slice...)
}

type Builder struct {
	parent *Builder

	m memory
	w *bytes.Buffer
}

// either returns dat if pointer is 0, or the 32 bytes of word
func (v *word) StackBytes() []byte {
	if v.pointer == 0 {
		return v.dat
	}
	o := [8]byte{}
	binary.BigEndian.PutUint64(o[:], uint64(v.pointer))
	return padLeft32(o[:])
}

type memory struct {
	stack []word

	// current stack pointer
	cur int

	heap [][]byte
	// current heap size
	hz int
}

func (m *memory) WriteStack(xs []byte) {
	xs = padRight32(xs)
	m.stack = append(m.stack, word{
		pointer: 0,
		dat:     xs,
	})
	m.cur = m.cur + len(xs)
	// increment existing pointers, which are now wrong since the initial offset has increased
	for k := range m.stack {
		if m.stack[k].pointer != 0 {
			m.stack[k].pointer = m.stack[k].pointer + len(xs)
		}
	}
}

func (m *memory) WriteDynamic(xs []byte) {
	xs = padRight32(xs)
	// add a chunk to the stack cursor
	m.cur = m.cur + 32
	// the current pointer is at the heap + current stack. add that to the stack
	m.stack = append(m.stack, word{
		pointer: m.hz + m.cur,
	})
	// now advance the heap cursor
	m.hz = m.hz + len(xs)
	// and append to heap
	m.heap = append(m.heap, xs)
}

func NewBuilder(xs []byte) *Builder {
	return NewBuilderRaw(bytes.NewBuffer(xs))
}

func NewBuilderRaw(w *bytes.Buffer) *Builder {
	return &Builder{
		w: w,
	}
}

func (d *Builder) WriteNPadRight32(xs []byte) *Builder {
	d.m.WriteStack(padRight32(xs))
	return d
}
func (d *Builder) WriteWord(xs []byte) {
	d.m.WriteStack(padLeft32(xs))
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

func (d *Builder) Close() error {
	for _, v := range d.m.stack {
		_, err := d.w.Write(v.StackBytes())
		if err != nil {
			return err
		}
	}
	for _, v := range d.m.heap {
		_, err := d.w.Write(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Builder) Finish() []byte {
	d.Close()
	return d.w.Bytes()
}

func (d *Builder) Reset() {
	d.w.Reset()
}

func (d *Builder) EnterDynamic(l int) *Builder {
	b := &Builder{
		parent: d,
	}
	b.WriteInt(l)
	return b
}
func (d *Builder) ExitDynamic() *Builder {
	if d.parent == nil {
		panic("tried to exit dynamic when not in one")
	}
	//TODO: a buf pool will reduce allocations here
	buf := new(bytes.Buffer)
	d.w = buf
	chile := d.Finish()
	d.parent.m.WriteDynamic(chile)
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
