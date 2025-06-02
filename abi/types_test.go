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

func TestUnslice(t *testing.T) {
	tn1, l1 := SLICE(UINT).UnSlice()
	assert.EqualValues(t, tn1, "uint256")
	assert.Equal(t, 0, l1)

	tn2, l2 := ARRAY(UINT, 14).UnSlice()
	assert.EqualValues(t, tn2, "uint256")
	assert.Equal(t, 14, l2)

	// nested
	tn3, l3 := SLICE(ARRAY(BYTES3, 3)).UnSlice()
	assert.EqualValues(t, tn3, "bytes3[3]")
	assert.Equal(t, 0, l3)

	tn4, l4 := ARRAY(SLICE(UINT), 4).UnSlice()
	assert.EqualValues(t, tn4, "uint256[]")
	assert.Equal(t, 4, l4)
}

func BenchmarkGenerateSignature(b *testing.B) {
	burn := SIG("burn", UINT256)
	for i := 0; i < b.N; i++ {
		burn.Hash()
	}
}

func TestTypeNameIsDynamicComprehensive(t *testing.T) {
	tests := []struct {
		input    TypeName
		expected bool
	}{
		// Dynamic base types
		{"string", true},
		{"bytes", true},

		// Static base types
		{"uint256", false},
		{"int256", false},
		{"address", false},
		{"bool", false},
		{"bytes32", false},

		// Dynamic slices
		{"uint256[]", true},
		{"address[]", true},
		{"bool[]", true},
		{"bytes32[]", true},

		// Fixed arrays of static types
		{"uint256[5]", false},
		{"address[10]", false},
		{"bool[3]", false},
		{"bytes32[2]", false},

		// Fixed arrays of dynamic types
		{"string[5]", true},
		{"bytes[10]", true},
		{"uint256[][5]", true}, // array of dynamic slices

		// Nested arrays
		{"uint256[][]", true},     // dynamic array of dynamic arrays
		{"uint256[5][]", true},    // dynamic array of fixed arrays
		{"uint256[][5]", true},    // fixed array of dynamic arrays
		{"uint256[5][10]", false}, // fixed array of fixed arrays
		{"string[5][10]", true},   // fixed array of fixed arrays of dynamic type

		// Tuples
		{"(uint256,address)", false},           // tuple of static types
		{"(uint256,string)", true},             // tuple with dynamic type
		{"(uint256,bytes)", true},              // tuple with dynamic type
		{"(uint256,uint256[])", true},          // tuple with dynamic array
		{"(uint256[5],address)", false},        // tuple of static types
		{"(uint256[],address)", true},          // tuple with dynamic array
		{"(uint256,(string,bytes))", true},     // nested tuple with dynamic
		{"(uint256,(uint256,address))", false}, // nested tuple all static
		{"((uint256,string),address)", true},   // nested tuple with dynamic

		// Complex cases
		{"(uint256,address)[5]", false},           // fixed array of dynamic tuples
		{"(uint256,string)[5]", true},             // fixed array of dynamic tuples
		{"(uint256,address)[]", true},             // dynamic array of static tuples
		{"(uint256,string)[]", true},              // dynamic array of dynamic tuples
		{"((uint256[],address),string[5])", true}, // complex nested dynamic
	}

	for _, test := range tests {
		result := test.input.IsDynamic()
		if result != test.expected {
			t.Errorf("TypeName(%q).IsDynamic() = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestTypeNameIsSlice(t *testing.T) {
	tests := []struct {
		input    TypeName
		expected bool
	}{
		{"uint256[]", true},
		{"uint256", false},
		{"bytes[]", true},
		{"bytes", false},
		{"uint256[5]", false},
		{"uint256[][]", true},
		{"(uint256,string)[]", true},
		{"(uint256,string)", false},
	}

	for _, test := range tests {
		result := test.input.IsSlice()
		if result != test.expected {
			t.Errorf("TypeName(%q).IsSlice() = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestTypeNameIsFixedSlice(t *testing.T) {
	tests := []struct {
		input    TypeName
		expected bool
	}{
		{"uint256[5]", true},
		{"uint256[]", false},
		{"uint256", false},
		{"bytes32[10]", true},
		{"uint256[5][10]", true},
		{"uint256[][10]", true},
		{"uint256[10][]", false},
		{"(uint256,string)[5]", true},
	}

	for _, test := range tests {
		result := test.input.IsFixedSlice()
		if result != test.expected {
			t.Errorf("TypeName(%q).IsFixedSlice() = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestTypeNameIsTuple(t *testing.T) {
	tests := []struct {
		input    TypeName
		expected bool
	}{
		{"(uint256,address)", true},
		{"(uint256)", true},
		{"()", true},
		{"uint256", false},
		{"(uint256,address)[]", false},
		{"((uint256,address),string)", true},
		{"[uint256,address]", false},
		{"(", false},
		{")", false},
	}

	for _, test := range tests {
		result := test.input.IsTuple()
		if result != test.expected {
			t.Errorf("TypeName(%q).IsTuple() = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestTypeNameTupleArgs(t *testing.T) {
	tests := []struct {
		input    TypeName
		expected []TypeName
	}{
		{"(uint256,address)", []TypeName{"uint256", "address"}},
		{"(uint256)", []TypeName{"uint256"}},
		{"(uint256,string,bool)", []TypeName{"uint256", "string", "bool"}},
		{"((uint256,address),string)", []TypeName{"(uint256,address)", "string"}},
		{"(uint256,(address,bool))", []TypeName{"uint256", "(address,bool)"}},
		{"(uint256[],address)", []TypeName{"uint256[]", "address"}},
		{"()", []TypeName{""}},
	}

	for _, test := range tests {
		result := test.input.TupleArgs()
		if len(result) != len(test.expected) {
			t.Errorf("TypeName(%q).TupleArgs() returned %d args, want %d", test.input, len(result), len(test.expected))
			continue
		}
		for i, arg := range result {
			if arg != test.expected[i] {
				t.Errorf("TypeName(%q).TupleArgs()[%d] = %q, want %q", test.input, i, arg, test.expected[i])
			}
		}
	}
}

func TestTypeNameIsValid(t *testing.T) {
	tests := []struct {
		input    TypeName
		expected bool
	}{
		// Valid base types
		{"bool", true},
		{"address", true},
		{"string", true},
		{"bytes", true},

		// Valid uint types
		{"uint8", true},
		{"uint16", true},
		{"uint24", true},
		{"uint32", true},
		{"uint256", true},

		// Valid int types
		{"int8", true},
		{"int16", true},
		{"int24", true},
		{"int32", true},
		{"int256", true},

		// Valid bytes types
		{"bytes1", true},
		{"bytes2", true},
		{"bytes17", true},
		{"bytes16", true},
		{"bytes32", true},

		// Valid fixed types
		{"fixed128x18", true},
		{"ufixed128x18", true},
		{"fixed256x80", true},
		{"ufixed8x0", true},

		// Valid arrays
		{"uint256[]", true},
		{"address[]", true},
		{"bool[]", true},
		{"string[]", true},
		{"uint256[5]", true},
		{"address[10]", true},
		{"bytes32[100]", true},

		// Valid nested arrays
		{"uint256[][]", true},
		{"uint256[5][10]", true},
		{"address[][5]", true},
		{"bool[10][]", true},

		// Valid tuples
		{"()", true},
		{"(uint256)", true},
		{"(uint256,address)", true},
		{"(uint256,string,bool)", true},
		{"((uint256,address),string)", true},
		{"(uint256,(address,bool))", true},

		// Valid complex types
		{"(uint256,address)[]", true},
		{"(uint256,string)[5]", true},
		{"(uint256[],address)", true},
		{"((uint256[],bool),address[])", true},

		// Invalid: empty
		{"", false},

		// Invalid: incomplete types
		{"uint", false},
		{"int", false},

		// Invalid: wrong sizes
		{"uint7", false},   // not multiple of 8
		{"uint257", false}, // > 256
		{"uint0", false},   // < 8
		{"int7", false},    // not multiple of 8
		{"int257", false},  // > 256
		{"int0", false},    // < 8
		{"bytes0", false},  // < 1
		{"bytes33", false}, // > 32
		{"bytes64", false}, // > 32

		// Invalid: fixed types
		{"fixed7x18", false},    // M not multiple of 8
		{"fixed256x81", false},  // N > 80
		{"fixed256x-1", false},  // N < 0
		{"ufixed300x18", false}, // M > 256
		{"fixed128x", false},    // missing N
		{"fixedx18", false},     // missing M
		{"fixed", false},        // missing MxN

		// Invalid: malformed types
		{"uint256[", false},
		{"uint256]", false},
		{"uint256[5", false},
		{"uint256]5[", false},
		{"[5]uint256", false},
		{"[]", false},
		{"[10]", false},

		// Invalid: bad array sizes
		{"uint256[0]", false},
		{"uint256[-1]", false},
		{"uint256[abc]", false},

		// Invalid: malformed tuples
		{"(", false},
		{")", false},
		{"(uint256,", false},
		{"uint256)", false},
		{"(uint256,,address)", false},

		// Invalid: nested invalid types
		{"uint7[]", false},
		{"(uint256,uint7)", false},
		{"(uint256,invalid)", false},
		{"invalid[]", false},
		{"invalid[5]", false},

		// Invalid: random strings
		{"hello", false},
		{"123", false},
		{"!@#$", false},
	}

	for _, test := range tests {
		result := test.input.IsValid()
		if result != test.expected {
			t.Errorf("TypeName(%q).IsValid() = %v, want %v", test.input, result, test.expected)
		}
	}
}
