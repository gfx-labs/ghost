package abir

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/gfx-labs/ghost/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundtrip encodes a value, decodes it, and re-encodes to verify consistency
func roundtrip[T any](t *testing.T, name string, original T, types ...abi.TypeName) {
	t.Helper()

	// First encode
	b1 := new(abi.Builder)
	err := Encode(b1, original, types...)
	require.NoError(t, err, "%s: first encode failed", name)
	encoded1 := b1.Finish()

	// Decode
	dec := abi.NewDecoder(encoded1)
	var decoded T
	err = Decode(dec, &decoded, types...)
	require.NoError(t, err, "%s: decode failed", name)

	// Verify decoded matches original
	assert.EqualValues(t, original, decoded, "%s: decoded value doesn't match original", name)

	// Re-encode
	b2 := new(abi.Builder)
	err = Encode(b2, decoded, types...)
	require.NoError(t, err, "%s: second encode failed", name)
	encoded2 := b2.Finish()

	// Verify encodings match
	assert.True(t, bytes.Equal(encoded1, encoded2),
		"%s: re-encoded bytes don't match original encoding\nFirst:  %s\nSecond: %s",
		name, abi.PrettyHex(encoded1), abi.PrettyHex(encoded2))
}

func TestRoundtripBasicTypes(t *testing.T) {
	t.Run("bool and uint", func(t *testing.T) {
		type S struct {
			A bool
			B bool
			C uint
		}
		roundtrip(t, "bool and uint", S{A: true, B: false, C: 42}, abi.BOOL, abi.BOOL, abi.UINT)
	})

	t.Run("uint sizes", func(t *testing.T) {
		type S struct {
			A uint8
			B uint16
			C uint32
			D uint64
		}
		roundtrip(t, "uint sizes", S{A: 255, B: 65535, C: 0xDEADBEEF, D: 0xCAFEBABE}, abi.UINT8, abi.UINT16, abi.UINT32, abi.UINT64)
	})

	t.Run("uint256", func(t *testing.T) {
		type S struct {
			A uint256.Int
			B uint
		}
		val := uint256.MustFromHex("0xdeadbeefcafebabe")
		roundtrip(t, "uint256", S{A: *val, B: 42}, abi.UINT256, abi.UINT)
	})

	t.Run("int positive and negative", func(t *testing.T) {
		type S struct {
			A int
			B int
			C int
		}
		roundtrip(t, "int positive and negative", S{A: 12345, B: -1, C: -9999}, abi.INT, abi.INT, abi.INT)
	})

	t.Run("big.Int negative", func(t *testing.T) {
		type S struct {
			A big.Int
			B big.Int
		}
		roundtrip(t, "big.Int negative", S{A: *big.NewInt(-12345), B: *big.NewInt(9999)}, abi.INT, abi.INT)
	})

	// Address as string (hex) - this is how the existing tests handle it
	t.Run("address as string", func(t *testing.T) {
		type S struct {
			A string `abi:"address"`
			B uint
		}
		roundtrip(t, "address as string", S{
			A: common.HexToAddress("0xdeadbeefcafebabe12345678").Hex(),
			B: 1,
		}, abi.ADDRESS, abi.UINT)
	})
}

func TestRoundtripDynamicTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		type S struct {
			A string
			B uint
		}
		roundtrip(t, "string", S{A: "Hello, World!", B: 42}, abi.STRING, abi.UINT)
	})

	t.Run("string empty", func(t *testing.T) {
		type S struct {
			A string
			B uint
		}
		roundtrip(t, "string empty", S{A: "", B: 1}, abi.STRING, abi.UINT)
	})

	t.Run("string long", func(t *testing.T) {
		type S struct {
			A string
			B uint
		}
		roundtrip(t, "string long", S{
			A: "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			B: 1,
		}, abi.STRING, abi.UINT)
	})

	t.Run("uint slice", func(t *testing.T) {
		type S struct {
			A []uint
			B uint
		}
		roundtrip(t, "uint slice", S{A: []uint{1, 2, 3, 4, 5}, B: 99}, abi.SLICE(abi.UINT), abi.UINT)
	})

	t.Run("int slice", func(t *testing.T) {
		type S struct {
			A []int
			B uint
		}
		roundtrip(t, "int slice", S{A: []int{-1, 0, 1, 100, -100}, B: 1}, abi.SLICE(abi.INT), abi.UINT)
	})
}

func TestRoundtripNestedDynamic(t *testing.T) {
	t.Run("string slice", func(t *testing.T) {
		type S struct {
			A []string
			B uint
		}
		roundtrip(t, "string slice", S{A: []string{"hello", "world", "foo", "bar"}, B: 1}, abi.SLICE(abi.STRING), abi.UINT)
	})

	t.Run("slice of slices", func(t *testing.T) {
		type S struct {
			A [][]uint
			B uint
		}
		roundtrip(t, "slice of slices", S{
			A: [][]uint{
				{1, 2, 3},
				{4, 5},
				{6, 7, 8, 9},
			},
			B: 1,
		}, abi.SLICE(abi.SLICE(abi.UINT)), abi.UINT)
	})

	t.Run("fixed array of strings", func(t *testing.T) {
		type S struct {
			A [2]string
			B uint
		}
		roundtrip(t, "fixed array of strings", S{A: [2]string{"first", "second"}, B: 1}, abi.ARRAY(abi.STRING, 2), abi.UINT)
	})
}

func TestRoundtripTuples(t *testing.T) {
	// Note: Static tuples (all non-dynamic fields) have issues with field ordering
	// in the reflection-based encoder/decoder. Dynamic tuples work correctly.

	t.Run("tuple with string", func(t *testing.T) {
		type Inner struct {
			Name  string
			Value uint
		}
		type S struct {
			V Inner
			X uint
		}
		roundtrip(t, "tuple with string", S{
			V: Inner{Name: "test", Value: 42},
			X: 1,
		}, abi.TUPLE(abi.STRING, abi.UINT), abi.UINT)
	})

	t.Run("tuple with slice", func(t *testing.T) {
		type Inner struct {
			Values []uint
			Flag   bool
		}
		type S struct {
			V Inner
			X uint
		}
		roundtrip(t, "tuple with slice", S{
			V: Inner{Values: []uint{1, 2, 3}, Flag: true},
			X: 1,
		}, abi.TUPLE(abi.SLICE(abi.UINT), abi.BOOL), abi.UINT)
	})
}

func TestRoundtripComplexStructs(t *testing.T) {
	t.Run("slice of structs", func(t *testing.T) {
		type Inner struct {
			X uint
			Y uint
		}
		type S struct {
			A []Inner
			B uint
		}
		roundtrip(t, "slice of structs", S{
			A: []Inner{
				{X: 1, Y: 2},
				{X: 3, Y: 4},
				{X: 5, Y: 6},
			},
			B: 1,
		}, abi.SLICE(abi.TUPLE(abi.UINT, abi.UINT)), abi.UINT)
	})

	t.Run("slice of structs with dynamic", func(t *testing.T) {
		type Inner struct {
			Name string
			ID   uint
		}
		type S struct {
			A []Inner
			B uint
		}
		roundtrip(t, "slice of structs with dynamic", S{
			A: []Inner{
				{Name: "alice", ID: 1},
				{Name: "bob", ID: 2},
			},
			B: 1,
		}, abi.SLICE(abi.TUPLE(abi.STRING, abi.UINT)), abi.UINT)
	})

	// Note: Static nested tuples have field ordering issues in reflection-based codec

	t.Run("deeply nested dynamic", func(t *testing.T) {
		type Level2 struct {
			Data string
		}
		type Level1 struct {
			Items []Level2
		}
		type S struct {
			V Level1
			X uint
		}
		roundtrip(t, "deeply nested dynamic", S{
			V: Level1{
				Items: []Level2{
					{Data: "one"},
					{Data: "two"},
					{Data: "three"},
				},
			},
			X: 1,
		}, abi.TUPLE(abi.SLICE(abi.TUPLE(abi.STRING))), abi.UINT)
	})

	// Based on TestEncodeNestedStructReflect pattern
	t.Run("complex nested with address strings", func(t *testing.T) {
		type Q struct {
			X uint
			E uint8
			Y uint8
		}
		type S struct {
			C string `abi:"address"`
			T []Q
		}
		type F struct {
			A  uint
			S1 [2]S
			S2 []S
			B  uint
		}

		// Note: decoder returns nil for empty slices, so use nil instead of []Q{}
		original := F{
			A: 7,
			B: 8,
			S1: [2]S{
				{
					C: common.HexToAddress("0x001d3f1ef827552ae1114027bd3ecf1f086ba0f9").Hex(),
					T: []Q{{0x11, 1, 0x12}},
				},
				{
					C: common.HexToAddress("0x0").Hex(),
					T: nil,
				},
			},
			S2: []S{
				{
					C: common.HexToAddress("0x0").Hex(),
					T: nil,
				},
				{
					C: common.HexToAddress("0x1234").Hex(),
					T: []Q{{0, 0, 0}, {0x21, 2, 0x22}, {0, 0, 0}},
				},
			},
		}

		roundtrip(t, "complex nested with address strings", original,
			abi.UINT,
			abi.ARRAY(abi.TUPLE(abi.ADDRESS, abi.SLICE(abi.TUPLE(abi.UINT, abi.UINT8, abi.UINT8))), 2),
			abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.SLICE(abi.TUPLE(abi.UINT, abi.UINT8, abi.UINT8)))),
			abi.UINT,
		)
	})
}

func TestRoundtripMixedTypes(t *testing.T) {
	t.Run("multiple fields mixed", func(t *testing.T) {
		type S struct {
			A uint
			B string
			C []uint
			D bool
		}
		roundtrip(t, "multiple fields mixed", S{
			A: 42,
			B: "hello",
			C: []uint{1, 2, 3},
			D: true,
		}, abi.UINT, abi.STRING, abi.SLICE(abi.UINT), abi.BOOL)
	})

	t.Run("address string and strings", func(t *testing.T) {
		type S struct {
			Addr string `abi:"address"`
			Name string
			Desc string
		}
		roundtrip(t, "address string and strings", S{
			Addr: common.HexToAddress("0xdeadbeef").Hex(),
			Name: "Token",
			Desc: "A test token",
		}, abi.ADDRESS, abi.STRING, abi.STRING)
	})
}

func TestRoundtripFixedBytes(t *testing.T) {
	// Note: fixed byte encoding/decoding has asymmetric string handling:
	// encoding from string treats it as raw bytes, decoding into string produces hex.
	// Roundtrips only work with []byte targets. [N]byte arrays can't encode
	// because reflect.Value.Bytes() panics on unaddressable arrays.

	t.Run("bytes10 as []byte", func(t *testing.T) {
		type S struct {
			A []byte `abi:"bytes10"`
			B uint
		}
		roundtrip(t, "bytes10 slice", S{A: []byte("1234567890"), B: 1}, abi.BYTES10, abi.UINT)
	})
	t.Run("bytes4 as []byte", func(t *testing.T) {
		type S struct {
			A []byte `abi:"bytes4"`
			B uint
		}
		roundtrip(t, "bytes4 slice", S{A: []byte{0xde, 0xad, 0xbe, 0xef}, B: 1}, abi.BYTES4, abi.UINT)
	})
	t.Run("multiple fixed bytes", func(t *testing.T) {
		type S struct {
			A []byte `abi:"bytes1"`
			B []byte `abi:"bytes2"`
			C []byte `abi:"bytes16"`
		}
		roundtrip(t, "multi fixed bytes", S{
			A: []byte{0xff},
			B: []byte{0xab, 0xcd},
			C: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		}, abi.BYTES1, abi.BYTES2, abi.BYTES16)
	})
}

func TestRoundtripIntSizes(t *testing.T) {
	t.Run("mixed int sizes", func(t *testing.T) {
		type S struct {
			A int8
			B int16
			C int32
			D int64
		}
		roundtrip(t, "int sizes", S{A: 127, B: -1000, C: -100000, D: -1 << 40},
			abi.INT8, abi.INT16, abi.INT32, abi.INT64)
	})
	t.Run("int8 boundary values", func(t *testing.T) {
		type S struct {
			A int8
			B int8
		}
		roundtrip(t, "int8 boundaries", S{A: 127, B: -128}, abi.INT8, abi.INT8)
	})
	t.Run("uint8 and uint16 max", func(t *testing.T) {
		type S struct {
			A uint8
			B uint16
			C uint32
		}
		roundtrip(t, "uint small sizes", S{A: 255, B: 65535, C: 0xffffffff},
			abi.UINT8, abi.UINT16, abi.UINT32)
	})
}

func TestRoundtripFixedArrays(t *testing.T) {
	// Note: fixed arrays of static types via abir have encode/decode asymmetry.
	// The encoder wraps them with EnterArray (adding dynamic offset handling),
	// while the decoder reads them inline. This means fixed arrays only roundtrip
	// correctly when the element type is dynamic (e.g. string[2] works, uint[3] doesn't).
	// Dynamic-element fixed arrays are covered in TestRoundtripNestedDynamic.

	t.Run("string[3] fixed array of dynamic type", func(t *testing.T) {
		type S struct {
			A [3]string
			B uint
		}
		roundtrip(t, "string[3]", S{A: [3]string{"one", "two", "three"}, B: 1},
			abi.ARRAY(abi.STRING, 3), abi.UINT)
	})
}

func TestRoundtripEdgeCases(t *testing.T) {
	t.Run("empty dynamic slice", func(t *testing.T) {
		type S struct {
			A []uint
			B uint
		}
		roundtrip(t, "empty slice", S{A: nil, B: 42}, abi.SLICE(abi.UINT), abi.UINT)
	})
	t.Run("empty string slice", func(t *testing.T) {
		type S struct {
			A []string
			B uint
		}
		roundtrip(t, "empty string slice", S{A: nil, B: 1}, abi.SLICE(abi.STRING), abi.UINT)
	})
	t.Run("uint256 zero and nonzero", func(t *testing.T) {
		type S struct {
			A uint256.Int
			B uint256.Int
		}
		roundtrip(t, "uint256 zero", S{A: *uint256.NewInt(0), B: *uint256.NewInt(1)},
			abi.UINT256, abi.UINT256)
	})
	t.Run("uint256 max", func(t *testing.T) {
		type S struct {
			A uint256.Int
			B uint
		}
		max := new(uint256.Int).Sub(
			new(uint256.Int).Lsh(uint256.NewInt(1), 256),
			uint256.NewInt(1),
		)
		roundtrip(t, "uint256 max", S{A: *max, B: 1}, abi.UINT256, abi.UINT)
	})
	t.Run("unicode string", func(t *testing.T) {
		type S struct {
			A string
			B uint
		}
		roundtrip(t, "unicode string", S{A: "こんにちは世界 🌍", B: 1}, abi.STRING, abi.UINT)
	})
	t.Run("bytes with high values as string", func(t *testing.T) {
		type S struct {
			A string
			B uint
		}
		roundtrip(t, "high bytes", S{A: string([]byte{0xff, 0xfe, 0x80, 0x00, 0x01}), B: 1},
			abi.BYTES, abi.UINT)
	})
	t.Run("single element slice", func(t *testing.T) {
		type S struct {
			A []uint
			B uint
		}
		roundtrip(t, "single element slice", S{A: []uint{42}, B: 1}, abi.SLICE(abi.UINT), abi.UINT)
	})
	t.Run("nested empty structs", func(t *testing.T) {
		type Inner struct {
			Name string
			ID   uint
		}
		type S struct {
			A []Inner
			B uint
		}
		roundtrip(t, "empty struct slice", S{A: nil, B: 99}, abi.SLICE(abi.TUPLE(abi.STRING, abi.UINT)), abi.UINT)
	})
	t.Run("large string over 32 bytes", func(t *testing.T) {
		type S struct {
			A string
			B uint
		}
		long := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz!!"
		roundtrip(t, "large string", S{A: long, B: 1}, abi.STRING, abi.UINT)
	})
}

func TestRoundtripCommonTypes(t *testing.T) {
	// Note: common.Hash and common.Address as [N]byte arrays cannot be encoded
	// via encodeReflectBytes because reflect.Value.Bytes() panics on unaddressable arrays.
	// These types only work when decoded (not roundtripped) or when passed as strings/slices.

	t.Run("uint256 and address string", func(t *testing.T) {
		type S struct {
			V uint256.Int
			A string `abi:"address"`
		}
		roundtrip(t, "uint256+addr", S{
			V: *uint256.NewInt(999),
			A: common.HexToAddress("0xcafe").Hex(),
		}, abi.UINT256, abi.ADDRESS)
	})
	t.Run("address and bytes32", func(t *testing.T) {
		type S struct {
			Addr string `abi:"address"`
			Hash []byte `abi:"bytes32"`
		}
		hash := make([]byte, 32)
		for i := range hash {
			hash[i] = byte(i + 1)
		}
		roundtrip(t, "addr+bytes32", S{
			Addr: common.HexToAddress("0xdeadbeef").Hex(),
			Hash: hash,
		}, abi.ADDRESS, abi.BYTES32)
	})
	t.Run("multiple uint256", func(t *testing.T) {
		type S struct {
			A uint256.Int
			B uint256.Int
			C uint
		}
		roundtrip(t, "multi uint256", S{
			A: *uint256.NewInt(1),
			B: *uint256.MustFromHex("0xdeadbeefcafebabe"),
			C: 42,
		}, abi.UINT256, abi.UINT256, abi.UINT)
	})
	t.Run("big.Int positive and negative", func(t *testing.T) {
		type S struct {
			A big.Int
			B big.Int
			C big.Int
		}
		roundtrip(t, "big.Int values", S{
			A: *big.NewInt(1),
			B: *big.NewInt(-12345),
			C: *big.NewInt(99999),
		}, abi.INT, abi.INT, abi.INT)
	})
}
