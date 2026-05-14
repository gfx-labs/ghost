package abi

import (
	"bytes"
	"io"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000045
0000000000000000000000000000000000000000000000000000000000000001`)
	type baz struct {
		x int
		y bool
	}
	var r baz
	var err error
	r.x, err = dec.Int()
	if err != nil {
		t.Fatal(err)
	}
	r.y, err = dec.Bool()
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, 69, r.x)
	assert.EqualValues(t, true, r.y)
}

func TestSimple(t *testing.T) {
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000023`)
	type f struct {
		a uint
		b []uint
	}
	var r f
	var err error
	r.a, err = dec.Uint()
	if err != nil {
		t.Fatal("r.a", err)
	}
	arr_b, err := dec.Dynamic()
	if err != nil {
		t.Fatal("r.b", err)
	}
	array_len, err := arr_b.Int()
	if err != nil {
		t.Fatal("r.b_len", err)
	}
	r.b = make([]uint, 0, array_len)
	for i := 0; i < array_len; i++ {
		val, err := arr_b.Uint()
		if err != nil {
			t.Fatal("r.b_inside", err)
		}
		r.b = append(r.b, uint(val))
	}
	assert.EqualValues(t, 7, r.a)
	assert.EqualValues(t, []uint{0x21, 0x22, 0x23}, r.b)
}

func TestDynamic(t *testing.T) {
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000123
0000000000000000000000000000000000000000000000000000000000000080
3132333435363738393000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000456
0000000000000000000000000000000000000000000000000000000000000789
000000000000000000000000000000000000000000000000000000000000000d
48656c6c6f2c20776f726c642100000000000000000000000000000000000000`)
	type f struct {
		a int
		b []uint32
		c []byte
		d string
	}
	var r f
	var err error
	r.a, err = dec.Int()
	if err != nil {
		t.Fatal("r.a", err)
	}
	arr_b, err := dec.Dynamic()
	if err != nil {
		t.Fatal("r.b", err)
	}
	array_len, err := arr_b.Int()
	if err != nil {
		t.Fatal("r.b_len", err)
	}
	r.b = make([]uint32, 0, array_len)
	for i := 0; i < array_len; i++ {
		val, err := arr_b.Uint()
		if err != nil {
			t.Fatal("r.b_inside", err)
			t.Fatal(err)
		}
		r.b = append(r.b, uint32(val))
	}
	r.c, err = dec.ReadNPadRight32(10)
	if err != nil {
		t.Fatal(err)
	}
	arr_d, err := dec.Dynamic()
	if err != nil {
		t.Fatal("arr_d", err)
	}
	str_len, err := arr_d.Int()
	if err != nil {
		t.Fatal("str_len", err)
	}
	bts, err := arr_d.ReadNPadRight32(str_len)
	if err != nil {
		t.Fatal("str", err)
	}
	r.d = string(bts)

	assert.EqualValues(t, 0x123, r.a)
	assert.EqualValues(t, []uint32{0x456, 0x789}, r.b)
	assert.EqualValues(t, []byte("1234567890"), r.c)
	assert.EqualValues(t, "Hello, world!", r.d)
}

func TestStructSimple(t *testing.T) {
	dec := HexDecoder(`
00000000000000000000000000000000000000000000000000000000000ff010
0000000000000000000000000000000000000000000000000000000000ff0002
6162636400000000000000000000000000000000000000000000000000000000`)
	type f struct {
		a int    `abi:"int256"`
		b uint   `abi:"uint256"`
		c []byte `abi:"bytes16"`
	}
	var r f
	var err error
	r.a, err = dec.Int()
	if err != nil {
		t.Fatal("r.a", err)
	}
	r.b, err = dec.Uint()
	if err != nil {
		t.Fatal("r.b", err)
	}
	r.c, err = dec.ReadNPadRight32(16)
	r.c = bytes.TrimRight(r.c, "\x00")
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, 0xff010, r.a)
	assert.EqualValues(t, 0xff0002, r.b)
	assert.EqualValues(t, []byte("abcd"), r.c)
}

func TestComplex(t *testing.T) {
	// 7, 0x60, 7 * 0x20,
	// // b
	// 3, 0x21, 0x22, 0x23,
	// // c
	// 0x40, 0x80,
	// 8, string("abcdefgh"),
	// 52, string("ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ")
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000060
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000023
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000080
0000000000000000000000000000000000000000000000000000000000000008
6162636465666768000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000034
4142434445464748494a4b4c4d4e4f505152535455565758595a414243444546
4748494a4b4c4d4e4f505152535455565758595a000000000000000000000000`)
	type f struct {
		a uint      `abi:"uint256"`
		b []uint    `abi:"uint256[]"`
		c [2]string `abi:"bytes[2]"`
	}
	var r f
	var err error
	r.a, err = dec.Uint()
	if err != nil {
		t.Fatal("r.a", err)
	}
	arr_b, err := dec.Dynamic()
	if err != nil {
		t.Fatal("r.b", err)
	}
	array_len, err := arr_b.Int()
	if err != nil {
		t.Fatal("r.b_len", err)
	}
	r.b = make([]uint, 0, array_len)
	for i := 0; i < array_len; i++ {
		val, err := arr_b.Uint()
		if err != nil {
			t.Fatal("r.b_inside", err)
			t.Fatal(err)
		}
		r.b = append(r.b, uint(val))
	}
	arr_c, err := dec.Dynamic()
	if err != nil {
		t.Fatal("r.c", err)
	}
	for i := 0; i < len(r.c); i++ {
		val, err := arr_c.DString()
		if err != nil {
			t.Fatal("r.c_inside", err)
			t.Fatal(err)
		}
		r.c[i] = val
	}
	assert.EqualValues(t, 7, r.a)
	assert.EqualValues(t, []uint{0x21, 0x22, 0x23}, r.b)
	assert.EqualValues(t, [2]string{"abcdefgh", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ"}, r.c)
}

func TestRemaining(t *testing.T) {
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000001
0000000000000000000000000000000000000000000000000000000000000002`)
	// Initially all bytes remain
	assert.Equal(t, 64, len(dec.Remaining()))

	// Read one word, 32 bytes remain
	_, err := dec.Uint256()
	assert.NoError(t, err)
	assert.Equal(t, 32, len(dec.Remaining()))

	// Read another word, 0 bytes remain
	_, err = dec.Uint256()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(dec.Remaining()))
}

func TestPeek(t *testing.T) {
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000042`)

	// Peek should not advance cursor
	buf := make([]byte, 4)
	n, err := dec.Peek(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{0, 0, 0, 0}, buf)

	// Cursor unchanged, can still read full word
	val, err := dec.Uint256()
	assert.NoError(t, err)
	assert.Equal(t, uint256.NewInt(0x42), val)

	// Peek past end should error
	buf2 := make([]byte, 4)
	_, err = dec.Peek(buf2)
	assert.ErrorIs(t, err, ErrUnexpectedEOF)
}

func TestPeekWord(t *testing.T) {
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000099`)

	// PeekWord should not advance cursor
	word, err := dec.PeekWord()
	assert.NoError(t, err)
	assert.Equal(t, byte(0x99), word[31])

	// Can still read the same word
	val, err := dec.Uint256()
	assert.NoError(t, err)
	assert.Equal(t, uint256.NewInt(0x99), val)

	// PeekWord on empty decoder should error
	_, err = dec.PeekWord()
	assert.ErrorIs(t, err, ErrUnexpectedEOF)
}

func TestPeekUint256(t *testing.T) {
	dec := HexDecoder(`
00000000000000000000000000000000000000000000000000000000000000ff`)

	// PeekUint256 should not advance cursor
	val, err := dec.PeekUint256()
	assert.NoError(t, err)
	assert.Equal(t, uint256.NewInt(0xff), val)

	// Can still read the same value
	val2, err := dec.Uint256()
	assert.NoError(t, err)
	assert.Equal(t, uint256.NewInt(0xff), val2)

	// PeekUint256 on empty decoder should error
	_, err = dec.PeekUint256()
	assert.ErrorIs(t, err, ErrUnexpectedEOF)
}

func TestBigIntNegative(t *testing.T) {
	// -1 in two's complement (all 1s)
	dec := HexDecoder(`
ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`)
	val, err := dec.BigInt()
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), val.Int64())

	// -2 in two's complement
	dec2 := HexDecoder(`
fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe`)
	val2, err := dec2.BigInt()
	assert.NoError(t, err)
	assert.Equal(t, int64(-2), val2.Int64())

	// Large negative: -256
	dec3 := HexDecoder(`
ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00`)
	val3, err := dec3.BigInt()
	assert.NoError(t, err)
	assert.Equal(t, int64(-256), val3.Int64())
}

func TestDynamicLength(t *testing.T) {
	// Encode: offset (0x40 = 64), some data, then at offset: length (3), items
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000040
000000000000000000000000000000000000000000000000000000000000002a
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000001
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000003`)

	arrDec, length, err := dec.DynamicLength()
	assert.NoError(t, err)
	assert.Equal(t, 3, length)

	// Read the array elements
	for i := 1; i <= 3; i++ {
		val, err := arrDec.Int()
		assert.NoError(t, err)
		assert.Equal(t, i, val)
	}

	// Original decoder can continue reading
	val, err := dec.Int()
	assert.NoError(t, err)
	assert.Equal(t, 0x2a, val)
}

func TestSeek(t *testing.T) {
	dec := HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000001
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000003`)

	// SeekStart
	pos, err := dec.Seek(32, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, int64(32), pos)
	val, _ := dec.Int()
	assert.Equal(t, 2, val)

	// SeekCurrent (now at 64, seek +32)
	pos, err = dec.Seek(32, io.SeekCurrent)
	assert.NoError(t, err)
	assert.Equal(t, int64(96), pos)

	// SeekEnd (seek back 32 from end)
	pos, err = dec.Seek(32, io.SeekEnd)
	assert.NoError(t, err)
	assert.Equal(t, int64(64), pos)
	val, _ = dec.Int()
	assert.Equal(t, 3, val)

	// SeekStart back to beginning
	dec.Seek(0, io.SeekStart)
	val, _ = dec.Int()
	assert.Equal(t, 1, val)

	// Invalid seek (negative position)
	_, err = dec.Seek(-100, io.SeekStart)
	assert.Error(t, err)

	// Invalid seek (past end)
	_, err = dec.Seek(1000, io.SeekStart)
	assert.Error(t, err)
}
