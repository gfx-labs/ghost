package abir

import (
	"io"
	"math/big"
	"testing"

	"gfx.cafe/open/ghost/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicTypeReflect(t *testing.T) {
	dec := abi.HexDecoder(`
0000000000000000000000000000000000000000000000000000000000000045
0000000000000000000000000000000000000000000000000000000000000001`)
	One := &big.Int{}
	var Two bool

	err := Decode(dec, One, abi.INT192)
	require.NoError(t, err)
	err = Decode(dec, &Two, abi.BOOL)
	require.NoError(t, err)

	assert.EqualValues(t, int64(69), One.Int64())
	assert.True(t, Two)
}
func TestSimpleStruct(t *testing.T) {
	dec := abi.HexDecoder(`0000000000000000000000000000000000000000000000000000000000000045`)

	var res struct {
		Data uint256.Int `abi:"uint256"`
	}
	err := DecodeInto(dec, &res)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, uint64(69), res.Data.Uint64())
}

func TestUintPtrDecodeInto(t *testing.T) {
	var s struct {
		Data *uint256.Int `abi:"uint256"`
	}
	dec := abi.HexDecoder(
		`0000000000000000000000000000000000000000000000000000000000000006`)
	res := uint256.NewInt(0)
	s.Data = res

	err := DecodeInto(dec, &s)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, "6", res.String())
}

func TestUintDecodeInto(t *testing.T) {
	dec := abi.HexDecoder(
		`0000000000000000000000000000000000000000000000000000000000000006`)
	var res uint256.Int
	err := DecodeInto(dec, &res)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, "6", res.String())
}

func TestBytesDecodeInto(t *testing.T) {
	dec := abi.HexDecoder(
		`0600000000000000000000000000000000000000000000000000000000000000`)
	var res [32]byte
	err := DecodeInto(dec, &res)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, [32]byte{6}, res)
	dec.Seek(0, io.SeekStart)
}

func TestHashDecodeInto(t *testing.T) {
	dec := abi.HexDecoder(
		`0600000000000000000000000000000000000000000000000000000000000000`)
	var res common.Hash
	err := DecodeInto(dec, &res)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, [32]byte{6}, res)
	dec.Seek(0, io.SeekStart)
}

func TestAddressReflect(t *testing.T) {
	dec := abi.HexDecoder(
		`0000000000000000000000000000000000000000000000000000000000001234`)
	var addr common.Address
	err := Decode(dec, &addr, abi.ADDRESS)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, common.HexToAddress("0x1234"), addr)
}

func TestDynamicReflect(t *testing.T) {
	hex := `
0000000000000000000000000000000000000000000000000000000000000123
0000000000000000000000000000000000000000000000000000000000000080
3132333435363738393000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000456
0000000000000000000000000000000000000000000000000000000000000789
000000000000000000000000000000000000000000000000000000000000000d
48656c6c6f2c20776f726c642100000000000000000000000000000000000000`
	type f struct {
		A big.Int `abi:"int64"`
		B []uint  `abi:"uint64[]"`
		C []byte  `abi:"bytes10"`
		D string  `abi:"string"`
	}
	var r f
	dec := abi.HexDecoder(hex)
	err := Decode(dec, &r, abi.INT64, abi.SLICE(abi.UINT64), abi.BYTES10, abi.STRING)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, int64(0x123), r.A.Int64())
	assert.EqualValues(t, []uint{0x456, 0x789}, r.B)
	assert.EqualValues(t, []byte("1234567890"), r.C)
	assert.EqualValues(t, "Hello, world!", r.D)

	var r2 f
	dec2 := abi.HexDecoder(hex)
	err2 := DecodeInto(dec2, &r2)
	if err2 != nil {
		t.Fatal(err2)
	}
	assert.EqualValues(t, int64(0x123), r2.A.Int64())
	assert.EqualValues(t, []uint{0x456, 0x789}, r2.B)
	assert.EqualValues(t, []byte("1234567890"), r2.C)
	assert.EqualValues(t, "Hello, world!", r2.D)
}

func TestSimpleReflect(t *testing.T) {
	hex := `
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000023`
	type f struct {
		A uint   `abi:"uint256"`
		B []uint `abi:"uint256[]"`
	}
	var r f
	dec := abi.HexDecoder(hex)
	err := Decode(dec, &r, abi.UINT, abi.SLICE(abi.UINT))
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, uint(7), r.A)
	assert.EqualValues(t, []uint{0x21, 0x22, 0x23}, r.B)
	var r2 f
	dec2 := abi.HexDecoder(hex)
	err2 := DecodeInto(dec2, &r2)
	if err2 != nil {
		t.Fatal(err2)
	}
	assert.EqualValues(t, uint(7), r2.A)
	assert.EqualValues(t, []uint{0x21, 0x22, 0x23}, r2.B)
}

func TestComplexReflect(t *testing.T) {
	// 7, 0x60, 7 * 0x20,
	// // b
	// 3, 0x21, 0x22, 0x23,
	// // c
	// 0x40, 0x80,
	// 8, string("abcdefgh"),
	// 52, string("ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hex := `
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
4748494a4b4c4d4e4f505152535455565758595a000000000000000000000000`
	type f struct {
		A uint      `abi:"uint256"`
		B []uint    `abi:"uint256[]"`
		C [2]string `abi:"bytes[2]"`
	}
	var r f
	dec := abi.HexDecoder(hex)
	err := Decode(dec, &r, abi.UINT, abi.SLICE(abi.UINT), abi.ARRAY(abi.BYTES, 2))
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, uint(7), r.A)
	assert.EqualValues(t, []uint{0x21, 0x22, 0x23}, r.B)
	assert.EqualValues(t, [2]string{"abcdefgh", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ"}, r.C)

	var r2 f
	dec2 := abi.HexDecoder(hex)
	err2 := DecodeInto(dec2, &r2)
	if err2 != nil {
		t.Fatal(err2)
	}
	assert.EqualValues(t, uint(7), r2.A)
	assert.EqualValues(t, []uint{0x21, 0x22, 0x23}, r2.B)
	assert.EqualValues(t, [2]string{"abcdefgh", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ"}, r2.C)
}

func TestStructComplex(t *testing.T) {
	// 7, 0x80, 0x1e0, 8,
	// 0x40,
	// 0x100,
	// m_contractAddress,
	// 0x40,
	// 1, // length
	// 0x11, 1, 0x12,
	// 0, 0x40,
	// 0,
	// 2, // length
	// 0x40, 0xa0,
	// 0, 0x40, 0,
	// 0x1234, 0x40,
	// 3, // length
	// 0, 0, 0,
	// 0x21, 2, 0x22,
	// 0, 0, 0
	hex := `
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000080
00000000000000000000000000000000000000000000000000000000000001e0
0000000000000000000000000000000000000000000000000000000000000008
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000100
000000000000000000000000001d3f1ef827552ae1114027bd3ecf1f086ba0f9
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000001
0000000000000000000000000000000000000000000000000000000000000011
0000000000000000000000000000000000000000000000000000000000000001
0000000000000000000000000000000000000000000000000000000000000012
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000040
00000000000000000000000000000000000000000000000000000000000000a0
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000001234
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000`
	type Q struct {
		X uint
		E uint8
		Y uint8
	}
	type S struct {
		C string `abi:"address"`
		T []Q
	}
	type f struct {
		A  uint
		S1 [2]S
		S2 []S
		B  uint
	}
	var r f
	dec := abi.HexDecoder(hex)
	err := Decode(dec, &r,
		abi.UINT,
		abi.ARRAY(abi.TUPLE(abi.ADDRESS, abi.SLICE(abi.TUPLE(abi.UINT, abi.UINT8, abi.UINT8))), 2),
		abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.SLICE(abi.TUPLE(abi.UINT, abi.UINT8, abi.UINT8)))),
		abi.UINT)
	if err != nil {
		t.Fatal(err)
	}
	// r.A
	assert.EqualValues(t, uint(7), r.A)
	// r.S1[0]
	assert.EqualValues(t, common.HexToAddress("0x001d3f1ef827552ae1114027bd3ecf1f086ba0f9").Hex(), r.S1[0].C)
	assert.EqualValues(t, Q{0x11, 1, 0x12}, r.S1[0].T[0])
	// r.S1[1]
	assert.EqualValues(t, common.HexToAddress("0x0").Hex(), r.S1[1].C)
	assert.Empty(t, r.S1[1].T)

	assert.EqualValues(t, 2, len(r.S2))
	// r.S2[0]
	assert.EqualValues(t, common.HexToAddress("0x0").Hex(), r.S2[0].C)
	assert.Empty(t, r.S2[0].T)
	// r.S2[1]
	assert.EqualValues(t, common.HexToAddress("0x1234").Hex(), r.S2[1].C)
	assert.EqualValues(t, 3, len(r.S2[1].T))
	assert.EqualValues(t, Q{0, 0, 0}, r.S2[1].T[0])
	assert.EqualValues(t, Q{0x21, 2, 0x22}, r.S2[1].T[1])
	assert.EqualValues(t, Q{0, 0, 0}, r.S2[1].T[2])
	// r.B
	assert.EqualValues(t, uint(8), r.B)

	var r2 f
	dec2 := abi.HexDecoder(hex)
	err2 := DecodeInto(dec2, &r2)
	if err2 != nil {
		t.Fatal(err2)
	}
	// r.A
	assert.EqualValues(t, uint(7), r2.A)
	// r.S1[0]
	assert.EqualValues(t, common.HexToAddress("0x001d3f1ef827552ae1114027bd3ecf1f086ba0f9").Hex(), r2.S1[0].C)
	assert.EqualValues(t, Q{0x11, 1, 0x12}, r2.S1[0].T[0])
	// r.S1[1]
	assert.EqualValues(t, common.HexToAddress("0x0").Hex(), r2.S1[1].C)
	assert.Empty(t, r2.S1[1].T)

	assert.EqualValues(t, 2, len(r2.S2))
	// r.S2[0]
	assert.EqualValues(t, common.HexToAddress("0x0").Hex(), r2.S2[0].C)
	assert.Empty(t, r2.S2[0].T)
	// r.S2[1]
	assert.EqualValues(t, common.HexToAddress("0x1234").Hex(), r2.S2[1].C)
	assert.EqualValues(t, 3, len(r2.S2[1].T))
	assert.EqualValues(t, Q{0, 0, 0}, r2.S2[1].T[0])
	assert.EqualValues(t, Q{0x21, 2, 0x22}, r2.S2[1].T[1])
	assert.EqualValues(t, Q{0, 0, 0}, r2.S2[1].T[2])
	// r.B
	assert.EqualValues(t, uint(8), r2.B)
}
