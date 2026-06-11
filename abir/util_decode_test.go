package abir

import (
	"math/big"
	"testing"

	"github.com/gfx-labs/ghost/abi"
	"github.com/gfx-labs/ghost/testutil"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeBytes(t *testing.T) {
	hex := testutil.HexBytes(`0000000000000000000000000000000000000000000000000000000000000045`)
	var v uint
	require.NoError(t, DecodeBytes(hex, &v, abi.UINT))
	assert.Equal(t, uint(0x45), v)
}

func TestDecodeNumericTargets(t *testing.T) {
	const word42 = `000000000000000000000000000000000000000000000000000000000000002a`

	t.Run("int sizes", func(t *testing.T) {
		cases := []struct {
			name string
			hex  string
			typ  abi.TypeName
			want int64
		}{
			{"int8", `0000000000000000000000000000000000000000000000000000000000000042`, abi.INT8, 0x42},
			{"int16", `0000000000000000000000000000000000000000000000000000000000000123`, abi.INT16, 0x123},
			{"int32", `0000000000000000000000000000000000000000000000000000000000001234`, abi.INT32, 0x1234},
			{"int64", `0000000000000000000000000000000000000000000000000000000012345678`, abi.INT64, 0x12345678},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				var v int64
				require.NoError(t, Decode(testutil.HexDecoder(tc.hex), &v, tc.typ))
				assert.Equal(t, tc.want, v)
			})
		}
	})

	t.Run("uint sizes", func(t *testing.T) {
		cases := []struct {
			name string
			hex  string
			typ  abi.TypeName
			want uint64
		}{
			{"uint8", `00000000000000000000000000000000000000000000000000000000000000ff`, abi.UINT8, 0xff},
			{"uint16", `000000000000000000000000000000000000000000000000000000000000ffff`, abi.UINT16, 0xffff},
			{"uint32", `00000000000000000000000000000000000000000000000000000000deadbeef`, abi.UINT32, 0xdeadbeef},
			{"uint64", `000000000000000000000000000000000000000000000000deadbeefcafebabe`, abi.UINT64, 0xdeadbeefcafebabe},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				var v uint64
				require.NoError(t, Decode(testutil.HexDecoder(tc.hex), &v, tc.typ))
				assert.Equal(t, tc.want, v)
			})
		}
	})

	t.Run("float64", func(t *testing.T) {
		var v float64
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		assert.Equal(t, float64(42), v)
	})
	t.Run("float32", func(t *testing.T) {
		var v float32
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		assert.Equal(t, float32(42), v)
	})
	t.Run("*big.Int", func(t *testing.T) {
		v := new(big.Int)
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		assert.Equal(t, int64(42), v.Int64())
	})
	t.Run("*uint256.Int", func(t *testing.T) {
		v := new(uint256.Int)
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		assert.Equal(t, uint64(42), v.Uint64())
	})
	t.Run("[]uint64 slice", func(t *testing.T) {
		var v []uint64
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		require.Len(t, v, 4)
		assert.Equal(t, uint64(42), v[0]) // low limb at index 0
	})
	t.Run("[]int64 slice", func(t *testing.T) {
		var v []int64
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		require.Len(t, v, 4)
		assert.Equal(t, int64(42), v[0])
	})
	t.Run("[]uint8 slice", func(t *testing.T) {
		var v []uint8
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		require.Len(t, v, 32)
		assert.Equal(t, uint8(42), v[0]) // big.Int.Bytes() big-endian
	})
	t.Run("[4]uint64 array", func(t *testing.T) {
		var v [4]uint64
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		assert.Equal(t, uint64(42), v[0])
	})
	t.Run("[4]int64 array", func(t *testing.T) {
		var v [4]int64
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		assert.Equal(t, int64(42), v[0])
	})
	t.Run("non-big.Int struct", func(t *testing.T) {
		type W struct{ Val uint64 }
		var v W
		require.NoError(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		assert.Equal(t, uint64(42), v.Val)
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("unsupported slice elem", func(t *testing.T) {
			var v []float64
			assert.Error(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		})
		t.Run("unsupported array elem", func(t *testing.T) {
			var v [4]float64
			assert.Error(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		})
		t.Run("unsupported kind", func(t *testing.T) {
			var v string
			assert.Error(t, Decode(testutil.HexDecoder(word42), &v, abi.UINT256))
		})
	})
}

func TestDecodeAddressTargets(t *testing.T) {
	const addrWord = `000000000000000000000000deadbeefcafebabedeadbeefcafebabedeadbeef`

	t.Run("into []byte", func(t *testing.T) {
		var v []byte
		require.NoError(t, Decode(testutil.HexDecoder(addrWord), &v, abi.ADDRESS))
		assert.Len(t, v, 20)
	})
	t.Run("into [20]byte", func(t *testing.T) {
		var v [20]byte
		require.NoError(t, Decode(testutil.HexDecoder(addrWord), &v, abi.ADDRESS))
		assert.Equal(t, byte(0xde), v[0])
	})
	t.Run("short array error", func(t *testing.T) {
		var v [10]byte
		assert.Error(t, Decode(testutil.HexDecoder(addrWord), &v, abi.ADDRESS))
	})
	t.Run("unsupported kind", func(t *testing.T) {
		var v int
		assert.Error(t, Decode(testutil.HexDecoder(addrWord), &v, abi.ADDRESS))
	})
}

func TestDecodeBoolTargets(t *testing.T) {
	const trueWord = `0000000000000000000000000000000000000000000000000000000000000001`
	const falseWord = `0000000000000000000000000000000000000000000000000000000000000000`

	t.Run("true into int", func(t *testing.T) {
		var v int
		require.NoError(t, Decode(testutil.HexDecoder(trueWord), &v, abi.BOOL))
		assert.Equal(t, 1, v)
	})
	t.Run("false into int", func(t *testing.T) {
		var v int
		require.NoError(t, Decode(testutil.HexDecoder(falseWord), &v, abi.BOOL))
		assert.Equal(t, 0, v)
	})
	t.Run("true into uint", func(t *testing.T) {
		var v uint
		require.NoError(t, Decode(testutil.HexDecoder(trueWord), &v, abi.BOOL))
		assert.Equal(t, uint(1), v)
	})
	t.Run("false into uint", func(t *testing.T) {
		var v uint
		require.NoError(t, Decode(testutil.HexDecoder(falseWord), &v, abi.BOOL))
		assert.Equal(t, uint(0), v)
	})
	t.Run("into string", func(t *testing.T) {
		var v string
		require.NoError(t, Decode(testutil.HexDecoder(trueWord), &v, abi.BOOL))
		assert.Equal(t, "true", v)
	})
}

func TestDecodeStringTargets(t *testing.T) {
	encodeStr := func(s string) []byte {
		b := new(abi.Builder)
		b.DString(s)
		return b.Finish()
	}

	t.Run("into []uint8", func(t *testing.T) {
		var v []uint8
		require.NoError(t, Decode(abi.NewDecoder(encodeStr("hi")), &v, abi.STRING))
		assert.Equal(t, []uint8{'h', 'i'}, v)
	})
	t.Run("into []int32", func(t *testing.T) {
		var v []int32
		require.NoError(t, Decode(abi.NewDecoder(encodeStr("AB")), &v, abi.STRING))
		assert.Equal(t, []int32{'A', 'B'}, v)
	})
	t.Run("into [10]uint8", func(t *testing.T) {
		var v [10]uint8
		require.NoError(t, Decode(abi.NewDecoder(encodeStr("hi")), &v, abi.STRING))
		assert.Equal(t, uint8('h'), v[0])
		assert.Equal(t, uint8('i'), v[1])
	})
	t.Run("short array error", func(t *testing.T) {
		var v [2]uint8
		assert.Error(t, Decode(abi.NewDecoder(encodeStr("hello")), &v, abi.STRING))
	})
	t.Run("unsupported array elem", func(t *testing.T) {
		var v [10]float64
		assert.Error(t, Decode(abi.NewDecoder(encodeStr("hi")), &v, abi.STRING))
	})
}

func TestDecodeFixedBytesTargets(t *testing.T) {
	const bytes4Word = `deadbeef00000000000000000000000000000000000000000000000000000000`

	t.Run("into string", func(t *testing.T) {
		var v string
		require.NoError(t, Decode(testutil.HexDecoder(bytes4Word), &v, abi.BYTES4))
		assert.Equal(t, "deadbeef", v)
	})
	t.Run("into []byte", func(t *testing.T) {
		var v []byte
		require.NoError(t, Decode(testutil.HexDecoder(bytes4Word), &v, abi.BYTES4))
		assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, v)
	})
	t.Run("into [4]int8", func(t *testing.T) {
		var v [4]int8
		require.NoError(t, Decode(testutil.HexDecoder(bytes4Word), &v, abi.BYTES4))
		assert.Equal(t, int8(-34), v[0]) // 0xde as int8
	})
	t.Run("unsupported array elem", func(t *testing.T) {
		var v [4]float64
		assert.Error(t, Decode(testutil.HexDecoder(bytes4Word), &v, abi.BYTES4))
	})
	t.Run("unsupported kind", func(t *testing.T) {
		var v int
		assert.Error(t, Decode(testutil.HexDecoder(bytes4Word), &v, abi.BYTES4))
	})
}

func TestDecodeDynamicBytesIntoByteSlice(t *testing.T) {
	input := []byte{0xde, 0xad, 0xbe, 0xef}
	b := new(abi.Builder)
	b.Bytes(input)

	var out []byte
	require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.BYTES))
	assert.Equal(t, input, out)
}

func TestEncodeNumericTypes(t *testing.T) {
	t.Run("[]byte as bytes", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, []byte{0xde, 0xad}, abi.BYTES))
		var out string
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.BYTES))
		assert.Equal(t, "\xde\xad", out)
	})
	t.Run("uint256.Int value", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, *uint256.NewInt(999), abi.UINT256))
		var out uint256.Int
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.UINT256))
		assert.Equal(t, uint64(999), out.Uint64())
	})
	t.Run("big.Int signed", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, *big.NewInt(-42), abi.INT256))
		var out big.Int
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.INT256))
		assert.Equal(t, int64(-42), out.Int64())
	})
	t.Run("uint256.Int for signed type", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, *uint256.NewInt(42), abi.INT256))
		var out uint256.Int
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.UINT256))
		assert.Equal(t, uint64(42), out.Uint64())
	})
	t.Run("big.Int for unsigned type", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, *big.NewInt(100), abi.UINT256))
		dec := abi.NewDecoder(b.Finish())
		u, err := dec.Uint256()
		require.NoError(t, err)
		assert.Equal(t, uint64(100), u.Uint64())
	})
}

func TestEncodeTypeErrors(t *testing.T) {
	t.Run("string into uint256", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), "hello", abi.UINT256))
	})
	t.Run("string into int256", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), "hello", abi.INT256))
	})
	t.Run("int into bytes4", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), 42, abi.BYTES4))
	})
	t.Run("int into address", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), 42, abi.ADDRESS))
	})
}
