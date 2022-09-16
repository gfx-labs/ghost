package abi

import (
	"testing"
)

func TestTupleType(t *testing.T) {
	TUPLE(UINT104, UINT120, UINT104)
}

func TestSigType(t *testing.T) {
	SIG("burn", UINT256)
}

func BenchmarkGenerateSignature(b *testing.B) {
	burn := SIG("burn", UINT256)
	for i := 0; i < b.N; i++ {
		burn.Hash()
	}
}
