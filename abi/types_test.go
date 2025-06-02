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
