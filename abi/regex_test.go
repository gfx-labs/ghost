package abi

import "testing"

func BenchmarkRegexFixedString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range testDataSet {
			fixedStringRegexCheck.MatchString(v)
		}
	}
}

func BenchmarkRegexDynamic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range testDataSet {
			dynamicRegexCheck.MatchString(v)
		}
	}
}
