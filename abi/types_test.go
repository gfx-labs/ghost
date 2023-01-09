package abi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSigBuilders(t *testing.T) {
	assert.EqualValues(t, TUPLE(UINT104, UINT120, UINT104), "(uint104,uint120,uint104)")
	assert.EqualValues(t, SIG("burn", UINT256), "burn(uint256)")
	assert.EqualValues(t, SLICE(TUPLE(UINT256, UINT256)), "(uint256,uint256)[]")
	assert.EqualValues(t, SLICE(SLICE(UINT256)), "uint256[][]")
	assert.EqualValues(t, ARRAY(UINT, 4), "uint256[4]")
}

func TestFixedArray(t *testing.T) {
	tn := ARRAY(UINT, 4)
	assert.False(t, tn.IsDynamic())
	tn = ARRAY(BYTES, 3)
	assert.True(t, tn.IsDynamic())
}

func TestIsDynamic(t *testing.T) {
	tn := BYTES
	assert.True(t, tn.IsDynamic())
	tn = SLICE(UINT)
	assert.True(t, tn.IsDynamic())
	tn = STRING
	assert.True(t, tn.IsDynamic())
	tn = TUPLE(INT, ARRAY(STRING, 5))
	assert.True(t, tn.IsDynamic())
	tn = TUPLE(ADDRESS, ARRAY(UINT, 3))
	assert.False(t, tn.IsDynamic())
}

func BenchmarkGenerateSignature(b *testing.B) {
	burn := SIG("burn", UINT256)
	for i := 0; i < b.N; i++ {
		burn.Hash()
	}
}
