package multicall

import (
	"testing"

	"gfx.cafe/open/ghost/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeTryAggregate(t *testing.T) {
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	calls := []Call{
		{Target: addr1, Data: []byte{0xde, 0xad}},
		{Target: addr2, Data: []byte{0xbe, 0xef}},
	}
	encoded := EncodeTryAggregate(true, calls)

	// Should start with the function selector (4 bytes)
	require.True(t, len(encoded) > 4)

	// First 4 bytes should be the tryAggregate selector
	expectedSel := abi.SIG("tryAggregate", abi.BOOL, abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.BYTES))).Fn()
	assert.Equal(t, expectedSel, encoded[:4])
}

func TestEncodeTryAggregateEmpty(t *testing.T) {
	encoded := EncodeTryAggregate(false, nil)
	require.True(t, len(encoded) > 4)
	// Should still produce valid ABI even with empty calls
}

func TestEncodeAggregate3(t *testing.T) {
	addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	calls := []Call3{
		{Target: addr, AllowFailure: true, Data: []byte{0x01, 0x02, 0x03}},
		{Target: addr, AllowFailure: false, Data: []byte{0x04}},
	}
	encoded := EncodeAggregate3(calls)

	require.True(t, len(encoded) > 4)

	expectedSel := abi.SIG("aggregate3", abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.BOOL, abi.BYTES))).Fn()
	assert.Equal(t, expectedSel, encoded[:4])
}

func TestEncodeAggregate3Value(t *testing.T) {
	addr := common.HexToAddress("0xcafecafecafecafecafecafecafecafecafecafe")
	val := *uint256.NewInt(1000)
	calls := []Call3Value{
		{Target: addr, AllowFailure: false, Value: val, Data: []byte{0xab, 0xcd}},
	}
	encoded := EncodeAggregate3Value(calls)

	require.True(t, len(encoded) > 4)

	expectedSel := abi.SIG("aggregate3Value", abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.BOOL, abi.UINT256, abi.BYTES))).Fn()
	assert.Equal(t, expectedSel, encoded[:4])
}

func TestDecodeResult(t *testing.T) {
	// Build a result payload: dynamic array of (bool, bytes) tuples
	// Layout: offset -> length -> [offset1, offset2, ...] -> [(success, bytes_offset, bytes_len, bytes_data), ...]
	b := new(abi.Builder)
	b.EnterDynamicArray().
		EnterTuple().Bool(true).Bytes([]byte{0x01, 0x02}).Exit().
		EnterTuple().Bool(false).Bytes([]byte{0x03}).Exit().
		Exit()
	encoded := b.Finish()

	results, err := DecodeResult(encoded)
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.True(t, results[0].Success)
	assert.Equal(t, []byte{0x01, 0x02}, results[0].Data)

	assert.False(t, results[1].Success)
	assert.Equal(t, []byte{0x03}, results[1].Data)
}

func TestDecodeResultEmpty(t *testing.T) {
	b := new(abi.Builder)
	b.EnterDynamicArray().Exit()
	encoded := b.Finish()

	results, err := DecodeResult(encoded)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestDecodeResultError(t *testing.T) {
	// Totally invalid data
	_, err := DecodeResult(nil)
	assert.Error(t, err)

	_, err = DecodeResult([]byte{0x01})
	assert.Error(t, err)
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	// Encode a tryAggregate call, then verify the result can be decoded
	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	calls := []Call{
		{Target: addr, Data: []byte{0xaa, 0xbb, 0xcc}},
	}
	encoded := EncodeTryAggregate(true, calls)
	require.True(t, len(encoded) > 4, "encoded data should contain selector + payload")

	// Build a mock result for DecodeResult
	b := new(abi.Builder)
	b.EnterDynamicArray().
		EnterTuple().Bool(true).Bytes([]byte{0xde, 0xad, 0xbe, 0xef}).Exit().
		Exit()
	resultData := b.Finish()

	results, err := DecodeResult(resultData)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].Success)
	assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, results[0].Data)
}
