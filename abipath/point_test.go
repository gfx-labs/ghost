package abipath

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/ghost/abi"
)

func word(v byte) []byte {
	var w [32]byte
	w[31] = v
	return w[:]
}

func words(vs ...byte) []byte {
	out := make([]byte, 0, len(vs)*32)
	for _, v := range vs {
		out = append(out, word(v)...)
	}
	return out
}

func padRight(s string) []byte {
	n := (len(s) + 31) / 32 * 32
	out := make([]byte, n)
	copy(out, s)
	return out
}

func TestPoint(t *testing.T) {
	// data layout: word(32)=offset to dynamic, word(64)=length, then payload
	data := bytes.Join([][]byte{
		word(64),     // offset 0: points to byte 64
		word(99),     // offset 32: filler
		word(3),      // offset 64: length = 3
		{1, 2, 3, 4}, // offset 96: payload
	}, nil)

	tests := []struct {
		name   string
		path   string
		input  []byte
		expect []byte
		err    bool
	}{
		{"empty path", "", data, data, false},
		{"dot dynamic", ".", data, data[64:], false},
		{"slash dynamic length", "/", data, data[96:], false},
		{"skip zero words", "0", data, data, false},
		{"skip one word", "1", data, data[32:], false},
		{"skip two words", "2", data, data[64:], false},
		{"skip to end", "3", data[:96], []byte{}, false},
		{"dot then skip", ".1", data, data[96:], false},
		{"slash then skip", "/0", data, data[96:], false},
		{"dot error", ".", word(255), nil, true},
		{"slash error", "/", word(32), nil, true},
		{"read error", "1", nil, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Point(tc.path, tc.input)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestPointErrors(t *testing.T) {
	huge := bytes.Repeat([]byte{0xff}, 32)

	tests := []struct {
		name  string
		path  string
		input []byte
		is    error
	}{
		{"skip past end", "2", words(1), abi.ErrUnexpectedEOF},
		{"skip on empty", "1", []byte{}, abi.ErrUnexpectedEOF},
		{"dot on empty", ".", nil, abi.ErrUnexpectedEOF},
		{"slash on empty", "/", nil, abi.ErrUnexpectedEOF},
		{"dot on short word", ".", make([]byte, 31), abi.ErrUnexpectedEOF},
		{"dot offset overflows uint64", ".", huge, nil},
		{"slash offset overflows uint64", "/", huge, nil},
		{"dot offset past end", ".", word(33), nil},
		{"slash offset past end", "/", word(33), nil},
		{"slash missing length word", "/", word(32), nil},
		{"error after valid step", "./", words(32, 64), nil},
		{"multi digit skip past end", "10", words(1, 2, 3, 4, 5, 6, 7, 8, 9), abi.ErrUnexpectedEOF},
		{"skip out of bounds", "2", words(1), ErrOutOfBounds},
		{"skip overflows int", "288230376151711744", nil, ErrOutOfBounds},
		{"skip exceeds atoi range", "99999999999999999999", words(1, 2), ErrOutOfBounds},
		{"huge skip does not allocate", "30000000", nil, ErrOutOfBounds},
		{"string", "foo", words(1), ErrInvalidToken},
		{"string before symbol", "foo/", words(1), ErrInvalidToken},
		{"string after integer", ".1b", words(32, 1, 2), ErrInvalidToken},
		{"string between symbols", "a.b1", words(32), ErrInvalidToken},
		{"space", "1 ", words(1, 2), ErrInvalidToken},
		{"non ascii", "é", words(1), ErrInvalidToken},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Point(tc.path, tc.input)
			require.Error(t, err)
			require.Nil(t, got)
			if tc.is != nil {
				require.ErrorIs(t, err, tc.is)
			}
		})
	}
}

func TestPointOffsetAtEnd(t *testing.T) {
	// an offset equal to the data length is valid and points at nothing
	got, err := Point(".", word(32))
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestPointMultiDigitSkip(t *testing.T) {
	data := words(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	got, err := Point("10", data)
	require.NoError(t, err)
	require.Equal(t, words(10, 11), got)
}

func TestPointAliasesInput(t *testing.T) {
	data := words(1, 2)
	got, err := Point("1", data)
	require.NoError(t, err)
	require.Equal(t, word(2), got)
	got[31] = 42
	require.Equal(t, byte(42), data[63])
}

// uint256[] a, int256 b
func TestPointDynamicArray(t *testing.T) {
	data := abi.NewBuilder().
		EnterDynamicArray().Int(123).Int(124).Exit().
		Int(4414).
		Finish()

	tests := []struct {
		name   string
		path   string
		expect []byte
	}{
		{"array head with length", ".", bytes.Join([][]byte{word(2), word(123), word(124)}, nil)},
		{"array elements", "/", bytes.Join([][]byte{word(123), word(124)}, nil)},
		{"second element", "/1", word(124)},
		{"past last element", "/2", []byte{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Point(tc.path, data)
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}

	t.Run("second param", func(t *testing.T) {
		got, err := Point("1", data)
		require.NoError(t, err)
		v, err := abi.NewDecoder(got).Int()
		require.NoError(t, err)
		require.Equal(t, 4414, v)
	})
}

// string s, int256 b
func TestPointString(t *testing.T) {
	data := abi.NewBuilder().
		DString("hello!").
		Int(4414).
		Finish()

	got, err := Point("/", data)
	require.NoError(t, err)
	require.Equal(t, padRight("hello!"), got)

	got, err = Point(".", data)
	require.NoError(t, err)
	require.Equal(t, append(word(6), padRight("hello!")...), got)

	// empty path leaves the decoder at the head, so the string decodes normally
	got, err = Point("", data)
	require.NoError(t, err)
	s, err := abi.NewDecoder(got).DString()
	require.NoError(t, err)
	require.Equal(t, "hello!", s)
}

// uint256[][] a, string[] b
func TestPointNested(t *testing.T) {
	data := abi.NewBuilder().
		EnterDynamicArray().
		EnterDynamicArray().Int(1).Int(2).Exit().
		EnterDynamicArray().Int(3).Exit().
		Exit().
		EnterDynamicArray().
		DString("one").DString("two").DString("three").
		Exit().
		Finish()

	tests := []struct {
		name   string
		path   string
		prefix []byte
	}{
		{"outer array length", ".", word(2)},
		{"first inner array", "//", words(1, 2)},
		{"first inner second element", "//1", word(2)},
		{"second inner array", "/1/", word(3)},
		{"second inner array head", "/1.", words(1, 3)},
		{"string array length", "1.", word(3)},
		{"first string", "1//", padRight("one")},
		{"second string", "1/1/", padRight("two")},
		{"third string", "1/2/", padRight("three")},
		{"third string head", "1/2.", append(word(5), padRight("three")...)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Point(tc.path, data)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(got), len(tc.prefix))
			require.Equal(t, tc.prefix, got[:len(tc.prefix)])
		})
	}

	t.Run("third string is at end of data", func(t *testing.T) {
		got, err := Point("1/2/", data)
		require.NoError(t, err)
		require.Equal(t, padRight("three"), got)
	})
}

func TestPointNoAlloc(t *testing.T) {
	// skipping must not copy, so large skips cost the same as none
	data := make([]byte, 1<<20)
	allocs := testing.AllocsPerRun(10, func() {
		_, _ = Point("32767", data)
	})
	require.LessOrEqual(t, allocs, float64(1))

	allocs = testing.AllocsPerRun(10, func() {
		_, _ = Point("30000000", nil)
	})
	require.LessOrEqual(t, allocs, float64(1))
}
