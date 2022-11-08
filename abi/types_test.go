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
}

func BenchmarkGenerateSignature(b *testing.B) {
	burn := SIG("burn", UINT256)
	for i := 0; i < b.N; i++ {
		burn.Hash()
	}
}
