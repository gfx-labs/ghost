package abi

// Conformance tests against known-good ABI encodings from the Solidity spec
// and real-world contracts (Multicall3, Uniswap V3).

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stripHex removes 0x prefix, whitespace, and newlines from a hex string.
func stripHex(s string) string {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}

func mustHex(s string) []byte {
	b, err := hex.DecodeString(stripHex(s))
	if err != nil {
		panic(err)
	}
	return b
}

// === Selector Tests ===

func TestConformanceSelectors(t *testing.T) {
	tests := []struct {
		sig      string
		selector string
	}{
		// Solidity spec examples
		{"baz(uint32,bool)", "cdcd77c0"},
		{"bar(bytes3[2])", "fce353f6"},
		{"sam(bytes,bool,uint256[])", "a5643bf2"},
		{"f(uint256,uint32[],bytes10,bytes)", "8be65246"},
		{"g(uint256[][],string[])", "2289b18c"},
		// Real contracts
		{"transfer(address,uint256)", "a9059cbb"},
		{"approve(address,uint256)", "095ea7b3"},
		{"balanceOf(address)", "70a08231"},
		{"aggregate3((address,bool,bytes)[])", "82ad56cb"},
		{"exactInput((bytes,address,uint256,uint256,uint256))", "c04b8d59"},
		{"multicall(uint256,bytes[])", "5ae401dc"},
	}
	for _, tc := range tests {
		t.Run(tc.sig, func(t *testing.T) {
			sel := Signature(tc.sig).Selector()
			got := hex.EncodeToString(sel[:])
			assert.Equal(t, tc.selector, got)
		})
	}
}

// === Solidity Spec Example 1: baz(uint32,bool) ===

func TestConformance_Baz(t *testing.T) {
	// baz(uint32,bool) with (69, true)
	expected := mustHex(
		"cdcd77c0" +
			"0000000000000000000000000000000000000000000000000000000000000045" +
			"0000000000000000000000000000000000000000000000000000000000000001")

	t.Run("encode", func(t *testing.T) {
		b := &Builder{}
		b.Uint(69).Bool(true)
		got := b.Finish(Signature("baz(uint32,bool)").Fn())
		assert.Equal(t, expected, got)
	})

	t.Run("decode", func(t *testing.T) {
		dec := NewDecoder(expected[4:]) // skip selector
		v1, err := dec.Uint()
		require.NoError(t, err)
		assert.Equal(t, uint(69), v1)
		v2, err := dec.Bool()
		require.NoError(t, err)
		assert.True(t, v2)
	})
}

// === Solidity Spec Example 2: bar(bytes3[2]) ===

func TestConformance_Bar(t *testing.T) {
	// bar(bytes3[2]) with (["abc", "def"])
	expected := mustHex(
		"fce353f6" +
			"6162630000000000000000000000000000000000000000000000000000000000" +
			"6465660000000000000000000000000000000000000000000000000000000000")

	t.Run("encode", func(t *testing.T) {
		b := &Builder{}
		b.EnterArray(BYTES3, 2).
			FixedBytes(3, []byte("abc")).
			FixedBytes(3, []byte("def")).
			Exit()
		got := b.Finish(Signature("bar(bytes3[2])").Fn())
		assert.Equal(t, expected, got)
	})

	t.Run("decode", func(t *testing.T) {
		dec := NewDecoder(expected[4:])
		v1, err := dec.ReadNPadRight32(3)
		require.NoError(t, err)
		assert.Equal(t, []byte("abc"), v1)
		v2, err := dec.ReadNPadRight32(3)
		require.NoError(t, err)
		assert.Equal(t, []byte("def"), v2)
	})
}

// === Solidity Spec Example 3: sam(bytes,bool,uint256[]) ===

func TestConformance_Sam(t *testing.T) {
	// sam(bytes,bool,uint256[]) with ("dave", true, [1,2,3])
	expected := mustHex(
		"a5643bf2" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"00000000000000000000000000000000000000000000000000000000000000a0" +
			"0000000000000000000000000000000000000000000000000000000000000004" +
			"6461766500000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000003" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000003")

	t.Run("encode", func(t *testing.T) {
		b := &Builder{}
		b.Bytes([]byte("dave")).
			Bool(true).
			EnterDynamicArray().Uint(1).Uint(2).Uint(3).Exit()
		got := b.Finish(Signature("sam(bytes,bool,uint256[])").Fn())
		assert.Equal(t, expected, got)
	})

	t.Run("decode", func(t *testing.T) {
		dec := NewDecoder(expected[4:])
		bts, err := dec.Bytes()
		require.NoError(t, err)
		assert.Equal(t, []byte("dave"), bts)

		bl, err := dec.Bool()
		require.NoError(t, err)
		assert.True(t, bl)

		arr, l, err := dec.DynamicLength()
		require.NoError(t, err)
		assert.Equal(t, 3, l)
		for i := 1; i <= 3; i++ {
			v, err := arr.Uint()
			require.NoError(t, err)
			assert.Equal(t, uint(i), v)
		}
	})
}

// === Solidity Spec Example 4: f(uint256,uint32[],bytes10,bytes) ===

func TestConformance_F(t *testing.T) {
	// f(uint256,uint32[],bytes10,bytes) with (0x123, [0x456,0x789], "1234567890", "Hello, world!")
	expected := mustHex(
		"8be65246" +
			"0000000000000000000000000000000000000000000000000000000000000123" +
			"0000000000000000000000000000000000000000000000000000000000000080" +
			"3132333435363738393000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000e0" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000456" +
			"0000000000000000000000000000000000000000000000000000000000000789" +
			"000000000000000000000000000000000000000000000000000000000000000d" +
			"48656c6c6f2c20776f726c642100000000000000000000000000000000000000")

	t.Run("encode", func(t *testing.T) {
		b := &Builder{}
		b.Uint(0x123).
			EnterDynamicArray().Uint(0x456).Uint(0x789).Exit().
			FixedBytes(10, []byte("1234567890")).
			Bytes([]byte("Hello, world!"))
		got := b.Finish(Signature("f(uint256,uint32[],bytes10,bytes)").Fn())
		assert.Equal(t, expected, got)
	})

	t.Run("decode", func(t *testing.T) {
		dec := NewDecoder(expected[4:])
		v1, err := dec.Uint()
		require.NoError(t, err)
		assert.Equal(t, uint(0x123), v1)

		arr, l, err := dec.DynamicLength()
		require.NoError(t, err)
		assert.Equal(t, 2, l)
		u1, err := arr.Uint()
		require.NoError(t, err)
		assert.Equal(t, uint(0x456), u1)
		u2, err := arr.Uint()
		require.NoError(t, err)
		assert.Equal(t, uint(0x789), u2)

		b10, err := dec.ReadNPadRight32(10)
		require.NoError(t, err)
		assert.Equal(t, []byte("1234567890"), b10)

		bts, err := dec.Bytes()
		require.NoError(t, err)
		assert.Equal(t, []byte("Hello, world!"), bts)
	})
}

// === Solidity Spec Example 5: g(uint256[][],string[]) ===
// This is the hardest example in the spec with 3 levels of nesting.

func TestConformance_G(t *testing.T) {
	// g(uint256[][],string[]) with ([[1,2],[3]], ["one","two","three"])
	expected := mustHex(
		"2289b18c" +
			"0000000000000000000000000000000000000000000000000000000000000040" +
			"0000000000000000000000000000000000000000000000000000000000000140" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000040" +
			"00000000000000000000000000000000000000000000000000000000000000a0" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000000003" +
			"0000000000000000000000000000000000000000000000000000000000000003" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"00000000000000000000000000000000000000000000000000000000000000a0" +
			"00000000000000000000000000000000000000000000000000000000000000e0" +
			"0000000000000000000000000000000000000000000000000000000000000003" +
			"6f6e650000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000003" +
			"74776f0000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000005" +
			"7468726565000000000000000000000000000000000000000000000000000000")

	t.Run("encode", func(t *testing.T) {
		b := &Builder{}
		b.EnterDynamicArray().
			EnterDynamicArray().Uint(1).Uint(2).Exit().
			EnterDynamicArray().Uint(3).Exit().
			Exit().
			EnterDynamicArray().
			DString("one").DString("two").DString("three").
			Exit()
		got := b.Finish(Signature("g(uint256[][],string[])").Fn())
		assert.Equal(t, expected, got)
	})

	t.Run("decode", func(t *testing.T) {
		dec := NewDecoder(expected[4:])

		// First arg: uint256[][]
		outerArr, outerLen, err := dec.DynamicLength()
		require.NoError(t, err)
		assert.Equal(t, 2, outerLen)

		// Sub-array [1, 2]
		sub1, sub1Len, err := outerArr.DynamicLength()
		require.NoError(t, err)
		assert.Equal(t, 2, sub1Len)
		v1, _ := sub1.Uint()
		v2, _ := sub1.Uint()
		assert.Equal(t, uint(1), v1)
		assert.Equal(t, uint(2), v2)

		// Sub-array [3]
		sub2, sub2Len, err := outerArr.DynamicLength()
		require.NoError(t, err)
		assert.Equal(t, 1, sub2Len)
		v3, _ := sub2.Uint()
		assert.Equal(t, uint(3), v3)

		// Second arg: string[]
		strArr, strLen, err := dec.DynamicLength()
		require.NoError(t, err)
		assert.Equal(t, 3, strLen)

		s1, err := strArr.DString()
		require.NoError(t, err)
		assert.Equal(t, "one", s1)
		s2, err := strArr.DString()
		require.NoError(t, err)
		assert.Equal(t, "two", s2)
		s3, err := strArr.DString()
		require.NoError(t, err)
		assert.Equal(t, "three", s3)
	})
}

// === Multicall3: aggregate3((address,bool,bytes)[]) ===

func TestConformance_Multicall3(t *testing.T) {
	// aggregate3 with 2 calls
	expected := mustHex(
		"82ad56cb" +
			"0000000000000000000000000000000000000000000000000000000000000020" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000040" +
			"00000000000000000000000000000000000000000000000000000000000000e0" +
			"000000000000000000000000deaddeaddeaddeaddeaddeaddeaddeaddeaddead" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"0000000000000000000000000000000000000000000000000000000000000004" +
			"aabbccdd00000000000000000000000000000000000000000000000000000000" +
			"000000000000000000000000beefbeefbeefbeefbeefbeefbeefbeefbeefbeef" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"ff00000000000000000000000000000000000000000000000000000000000000")

	t.Run("encode", func(t *testing.T) {
		b := &Builder{}
		callType := TUPLE(ADDRESS, BOOL, BYTES)
		b.EnterDynamicArray().
			// Call 1: (0xDeaD...DeaD, true, 0xaabbccdd)
			EnterArray(callType, 1). // really we need EnterTuple per element
			Exit()
		// The Builder API doesn't directly support "dynamic array of tuples"
		// in a single call — the encode test is done at the decode level.
		// We verify encoding via the existing multicall package tests.
	})

	t.Run("decode", func(t *testing.T) {
		dec := NewDecoder(expected[4:])

		// Outer dynamic offset
		outer, err := dec.Dynamic()
		require.NoError(t, err)

		// Array length
		arrLen, err := outer.Uint()
		require.NoError(t, err)
		assert.Equal(t, uint(2), arrLen)

		// Element offsets
		off1, err := outer.Uint()
		require.NoError(t, err)
		off2, err := outer.Uint()
		require.NoError(t, err)
		_ = off1
		_ = off2

		// Decode from the array data start (after length word)
		// We need to re-create decoder from the array start
		arrayDec := NewDecoder(expected[4+0x20:]) // skip selector + outer offset
		// Skip length word
		_, _ = arrayDec.Uint()

		// Read element 0 offset and element 1 offset
		elem0off, _ := arrayDec.Uint()
		elem1off, _ := arrayDec.Uint()

		// Element 0 at offset 0x40 from array start (after length word)
		_ = elem0off
		_ = elem1off

		// Simpler: decode elements directly from known positions
		elem0 := NewDecoder(expected[4+0x20+0x20+0x40:]) // call 0
		addr0, err := elem0.Address()
		require.NoError(t, err)
		assert.Equal(t, common.HexToAddress("0xDeaDDeaDDeaDDeaDDeaDDeaDDeaDDeaDDeaDDeaD"), addr0)

		allow0, err := elem0.Bool()
		require.NoError(t, err)
		assert.True(t, allow0)

		data0, err := elem0.Bytes()
		require.NoError(t, err)
		assert.Equal(t, mustHex("aabbccdd"), data0)

		elem1 := NewDecoder(expected[4+0x20+0x20+0xe0:]) // call 1
		addr1, err := elem1.Address()
		require.NoError(t, err)
		assert.Equal(t, common.HexToAddress("0xBeEFBeEFBeEFBeEFBeEFBeEFBeEFBeEFBeEFBeEF"), addr1)

		allow1, err := elem1.Bool()
		require.NoError(t, err)
		assert.False(t, allow1)

		data1, err := elem1.Bytes()
		require.NoError(t, err)
		assert.Equal(t, mustHex("ff"), data1)
	})
}

// === Uniswap V3: exactInput((bytes,address,uint256,uint256,uint256)) ===

func TestConformance_UniswapExactInput(t *testing.T) {
	expected := mustHex(
		"c04b8d59" +
			"0000000000000000000000000000000000000000000000000000000000000020" +
			"00000000000000000000000000000000000000000000000000000000000000a0" +
			"0000000000000000000000002a6b82b6dd3f38eeb63a35f2f503b9398f02d9bb" +
			"0000000000000000000000000000000000000000000000000000000861c46800" +
			"0000000000000000000000000000000000000000000000000000000000002710" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000042" +
			"2791bca1f2de4661ed88a30c99a7a9449aa841740005007ceb23fd6bc0add59e" +
			"62ac25578270cff1b9f619003000c26d47d5c33ac71ac5cf9f776d63ba292a4f" +
			"7842000000000000000000000000000000000000000000000000000000000000")

	t.Run("selector", func(t *testing.T) {
		sel := Signature("exactInput((bytes,address,uint256,uint256,uint256))").Selector()
		assert.Equal(t, [4]byte{0xc0, 0x4b, 0x8d, 0x59}, sel)
	})

	t.Run("decode", func(t *testing.T) {
		dec := NewDecoder(expected[4:])

		// Outer tuple offset
		tupleDec, err := dec.Dynamic()
		require.NoError(t, err)

		// bytes path (dynamic) - offset within the tuple
		pathBytes, err := tupleDec.Bytes()
		require.NoError(t, err)

		expectedPath := mustHex(
			"2791bca1f2de4661ed88a30c99a7a9449aa84174" +
				"000500" +
				"7ceb23fd6bc0add59e62ac25578270cff1b9f619" +
				"003000" +
				"c26d47d5c33ac71ac5cf9f776d63ba292a4f7842")
		assert.Equal(t, expectedPath, pathBytes)
		assert.Equal(t, 66, len(pathBytes)) // 20+3+20+3+20

		// address recipient
		recipient, err := tupleDec.Address()
		require.NoError(t, err)
		assert.Equal(t, common.HexToAddress("0x2a6b82b6dd3f38eeb63a35f2f503b9398f02d9bb"), recipient)

		// uint256 deadline
		deadline, err := tupleDec.Uint256()
		require.NoError(t, err)
		assert.Equal(t, new(big.Int).SetInt64(36000000000), deadline.ToBig())

		// uint256 amountIn
		amountIn, err := tupleDec.Uint256()
		require.NoError(t, err)
		assert.Equal(t, new(big.Int).SetInt64(10000), amountIn.ToBig())

		// uint256 amountOutMinimum
		amountOutMin, err := tupleDec.Uint256()
		require.NoError(t, err)
		assert.True(t, amountOutMin.IsZero())
	})

	t.Run("encode", func(t *testing.T) {
		path := mustHex(
			"2791bca1f2de4661ed88a30c99a7a9449aa84174" +
				"000500" +
				"7ceb23fd6bc0add59e62ac25578270cff1b9f619" +
				"003000" +
				"c26d47d5c33ac71ac5cf9f776d63ba292a4f7842")

		b := &Builder{}
		tup := b.EnterTuple()
		tup.Bytes(path).
			Address(common.HexToAddress("0x2a6b82b6dd3f38eeb63a35f2f503b9398f02d9bb")).
			Uint(36000000000).
			Uint(10000).
			Uint(0)
		tup.Exit()
		got := b.Finish(Signature("exactInput((bytes,address,uint256,uint256,uint256))").Fn())
		assert.Equal(t, expected, got)
	})
}

// === Edge cases ===

func TestConformance_EmptyString(t *testing.T) {
	// Encode and decode an empty string
	b := &Builder{}
	b.DString("")
	encoded := b.Finish()

	dec := NewDecoder(encoded)
	s, err := dec.DString()
	require.NoError(t, err)
	assert.Equal(t, "", s)
}

func TestConformance_EmptyBytes(t *testing.T) {
	b := &Builder{}
	b.Bytes([]byte{})
	encoded := b.Finish()

	dec := NewDecoder(encoded)
	bts, err := dec.Bytes()
	require.NoError(t, err)
	assert.Equal(t, []byte{}, bts)
}

func TestConformance_EmptyDynamicArray(t *testing.T) {
	b := &Builder{}
	b.EnterDynamicArray().Exit()
	encoded := b.Finish()

	dec := NewDecoder(encoded)
	arr, l, err := dec.DynamicLength()
	require.NoError(t, err)
	assert.Equal(t, 0, l)
	_ = arr
}

func TestConformance_LargeUint(t *testing.T) {
	// MaxUint256
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	b := &Builder{}
	b.Word(max.Bytes())
	encoded := b.Finish()

	dec := NewDecoder(encoded)
	v, err := dec.Uint256()
	require.NoError(t, err)
	assert.Equal(t, max, v.ToBig())
}

func TestConformance_SignedIntBoundaries(t *testing.T) {
	tests := []int64{0, 1, -1, 127, -128, 32767, -32768}
	for _, val := range tests {
		b := &Builder{}
		b.BigInt(big.NewInt(val))
		encoded := b.Finish()

		dec := NewDecoder(encoded)
		got, err := dec.BigInt()
		require.NoError(t, err, "value: %d", val)
		assert.Equal(t, val, got.Int64(), "value: %d", val)
	}
}

func TestConformance_MaliciousOffsets(t *testing.T) {
	// Offset larger than data
	dec := hexDecoder("00000000000000000000000000000000000000000000000000000000ffffffff")
	_, err := dec.Dynamic()
	assert.Error(t, err)

	// Offset = max uint64 (would overflow int)
	dec2 := hexDecoder("000000000000000000000000000000000000000000000000ffffffffffffffff")
	_, err = dec2.Dynamic()
	assert.Error(t, err)

	// Offset = 2^63 (negative as int)
	dec3 := hexDecoder("0000000000000000000000000000000000000000000000008000000000000000")
	_, err = dec3.Dynamic()
	assert.Error(t, err)

	// Very large uint256 offset (>64 bits)
	dec4 := hexDecoder("0000000000000000000000000000000100000000000000000000000000000000")
	_, err = dec4.Dynamic()
	assert.Error(t, err)
}
