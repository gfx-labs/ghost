package abir

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/gfx-labs/ghost/abi"
	"github.com/gfx-labs/ghost/testutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const oneWord = `0000000000000000000000000000000000000000000000000000000000000001`

func TestDecodeErrors(t *testing.T) {
	t.Run("non-ptr DecodeInto", func(t *testing.T) {
		var v uint
		assert.Error(t, DecodeInto(testutil.HexDecoder(oneWord), v))
	})
	t.Run("non-settable", func(t *testing.T) {
		v := uint(0)
		assert.Error(t, Decode(testutil.HexDecoder(oneWord), v, abi.UINT))
	})
	t.Run("zero args", func(t *testing.T) {
		var v uint
		assert.Error(t, Decode(testutil.HexDecoder(oneWord), &v))
	})
	t.Run("multi-arg into non-struct", func(t *testing.T) {
		var v uint
		assert.Error(t, Decode(testutil.HexDecoder(oneWord), &v, abi.UINT, abi.UINT))
	})
	t.Run("unknown type", func(t *testing.T) {
		var v uint
		assert.Error(t, Decode(testutil.HexDecoder(oneWord), &v, abi.TypeName("foobar")))
	})
	t.Run("tuple into non-struct", func(t *testing.T) {
		var v uint
		assert.Error(t, Decode(testutil.HexDecoder(oneWord), &v, abi.TUPLE(abi.UINT)))
	})
	t.Run("slice into non-slice", func(t *testing.T) {
		dec := testutil.HexDecoder(`
			0000000000000000000000000000000000000000000000000000000000000020
			0000000000000000000000000000000000000000000000000000000000000001
			0000000000000000000000000000000000000000000000000000000000000042`)
		var v uint
		assert.Error(t, Decode(dec, &v, abi.SLICE(abi.UINT)))
	})
	t.Run("fixed array into non-array", func(t *testing.T) {
		var v uint
		assert.Error(t, Decode(testutil.HexDecoder(oneWord), &v, abi.ARRAY(abi.UINT, 1)))
	})
	t.Run("fixed array length mismatch", func(t *testing.T) {
		dec := testutil.HexDecoder(oneWord + oneWord)
		var v [3]uint
		assert.Error(t, Decode(dec, &v, abi.ARRAY(abi.UINT, 2)))
	})
	t.Run("panic recovery on truncated data", func(t *testing.T) {
		var v uint
		assert.Error(t, Decode(abi.NewDecoder([]byte{0x01}), &v, abi.UINT))
	})
}

func TestDecodeSkipTag(t *testing.T) {
	dec := testutil.HexDecoder(`
		0000000000000000000000000000000000000000000000000000000000000042
		0000000000000000000000000000000000000000000000000000000000000063`)
	type S struct {
		Skip uint `abi:"-"`
		A    uint
		B    uint
	}
	var s S
	require.NoError(t, Decode(dec, &s, abi.UINT, abi.UINT))
	assert.Equal(t, uint(0), s.Skip)
	assert.Equal(t, uint(0x42), s.A)
	assert.Equal(t, uint(0x63), s.B)
}

func TestEncodeErrors(t *testing.T) {
	t.Run("zero args", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), uint(1)))
	})
	t.Run("multi-arg into non-struct", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), uint(1), abi.UINT, abi.UINT))
	})
	t.Run("tuple into non-struct", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), uint(1), abi.TUPLE(abi.UINT)))
	})
	t.Run("fixed array into non-array", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), uint(1), abi.ARRAY(abi.UINT, 1)))
	})
	t.Run("fixed array length mismatch", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), [2]uint{1, 2}, abi.ARRAY(abi.UINT, 3)))
	})
	t.Run("slice into non-slice", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), uint(1), abi.SLICE(abi.UINT)))
	})
	t.Run("unknown type", func(t *testing.T) {
		assert.Error(t, Encode(new(abi.Builder), uint(1), abi.TypeName("foobar")))
	})
	t.Run("panic recovery", func(t *testing.T) {
		// FixedBytes with short input — caught by recover
		_ = Encode(new(abi.Builder), "ab", abi.BYTES3)
	})
	t.Run("EncodeArray length mismatch", func(t *testing.T) {
		v := reflect.ValueOf([2]uint{1, 2})
		assert.Error(t, EncodeArray(new(abi.Builder), abi.UINT, 3, v))
	})
}

func TestCreateTypeName(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		expected abi.TypeName
	}{
		{"uint256.Int", reflect.TypeOf(uint256.Int{}), abi.UINT256},
		{"*uint256.Int", reflect.TypeOf((*uint256.Int)(nil)), abi.UINT256},
		{"big.Int", reflect.TypeOf(big.Int{}), abi.UINT256},
		{"*big.Int", reflect.TypeOf((*big.Int)(nil)), abi.UINT256},
		{"common.Hash", reflect.TypeOf(common.Hash{}), abi.BYTES32},
		{"*common.Hash", reflect.TypeOf((*common.Hash)(nil)), abi.BYTES32},
		{"common.Address", reflect.TypeOf(common.Address{}), abi.ADDRESS},
		{"*common.Address", reflect.TypeOf((*common.Address)(nil)), abi.ADDRESS},
		{"bool", reflect.TypeOf(false), abi.BOOL},
		{"string", reflect.TypeOf(""), abi.STRING},
		{"uint", reflect.TypeOf(uint(0)), "uint"},
		{"uint8", reflect.TypeOf(uint8(0)), "uint8"},
		{"uint16", reflect.TypeOf(uint16(0)), "uint16"},
		{"uint32", reflect.TypeOf(uint32(0)), "uint32"},
		{"uint64", reflect.TypeOf(uint64(0)), "uint64"},
		{"int", reflect.TypeOf(int(0)), "int"},
		{"int8", reflect.TypeOf(int8(0)), "int8"},
		{"int64", reflect.TypeOf(int64(0)), "int64"},
		{"[]uint", reflect.TypeOf([]uint{}), abi.SLICE("uint")},
		{"[3]uint", reflect.TypeOf([3]uint{}), abi.ARRAY("uint", 3)},
		{"[20]byte", reflect.TypeOf([20]byte{}), "bytes20"},
		{"[32]byte", reflect.TypeOf([32]byte{}), "bytes32"},
		{"*uint", reflect.TypeOf((*uint)(nil)), "uint"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CreateTypeName(tc.typ))
		})
	}

	t.Run("struct", func(t *testing.T) {
		type S struct{ A uint; B string }
		assert.Equal(t, abi.TUPLE("uint", abi.STRING), CreateTypeName(reflect.TypeOf(S{})))
	})
	t.Run("struct with abi tag", func(t *testing.T) {
		type S struct{ A uint `abi:"uint256"`; B string `abi:"address"` }
		assert.Equal(t, abi.TUPLE(abi.UINT256, abi.ADDRESS), CreateTypeName(reflect.TypeOf(S{})))
	})
	t.Run("struct with skip tag", func(t *testing.T) {
		type S struct{ Skip uint `abi:"-"`; A uint }
		assert.Equal(t, abi.TUPLE("uint"), CreateTypeName(reflect.TypeOf(S{})))
	})
}

func TestDecodeInto(t *testing.T) {
	t.Run("single value", func(t *testing.T) {
		var v uint
		require.NoError(t, DecodeInto(testutil.HexDecoder(`0000000000000000000000000000000000000000000000000000000000000042`), &v))
		assert.Equal(t, uint(0x42), v)
	})
	t.Run("pointer", func(t *testing.T) {
		v := new(uint)
		require.NoError(t, DecodeInto(testutil.HexDecoder(`0000000000000000000000000000000000000000000000000000000000000042`), v))
		assert.Equal(t, uint(0x42), *v)
	})
}

func TestEncodeDecodeRoundtrips(t *testing.T) {
	t.Run("address as string", func(t *testing.T) {
		addr := common.HexToAddress("0xdeadbeefcafebabedeadbeefcafebabedeadbeef").Hex()
		b := new(abi.Builder)
		require.NoError(t, Encode(b, addr, abi.ADDRESS))
		var out string
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.ADDRESS))
		assert.Equal(t, addr, out)
	})
	t.Run("bool true", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, true, abi.BOOL))
		var out bool
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.BOOL))
		assert.True(t, out)
	})
	t.Run("bool false", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, false, abi.BOOL))
		var out bool
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.BOOL))
		assert.False(t, out)
	})
	t.Run("bytes as string", func(t *testing.T) {
		b := new(abi.Builder)
		require.NoError(t, Encode(b, "deadbeefcafe", abi.BYTES))
		var out string
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.BYTES))
		assert.Equal(t, "deadbeefcafe", out)
	})
	t.Run("uint256.Int", func(t *testing.T) {
		val := *uint256.NewInt(12345)
		b := new(abi.Builder)
		require.NoError(t, Encode(b, val, abi.UINT256))
		var out uint256.Int
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.UINT256))
		assert.Equal(t, val, out)
	})
	t.Run("big.Int negative", func(t *testing.T) {
		val := *big.NewInt(-999)
		b := new(abi.Builder)
		require.NoError(t, Encode(b, val, abi.INT256))
		var out big.Int
		require.NoError(t, Decode(abi.NewDecoder(b.Finish()), &out, abi.INT256))
		assert.Equal(t, int64(-999), out.Int64())
	})
}
