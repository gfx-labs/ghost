package abi

import "testing"

func TestIsFixedArray(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"uint256[5]", true},
		{"bytes32[10]", true},
		{"address[100]", true},
		{"uint256[]", false},
		{"string", false},
		{"uint256[5][10]", true}, // nested fixed arrays
		{"uint256[][10]", true},  // dynamic array with fixed outer
		{"uint256[10][]", false}, // fixed array with dynamic outer - ends with []
		{"[10]", true},
		{"[]", false},
		{"[", false},
		{"]", false},
		{"[a]", false},
		{"[1a]", false},
		{"[10", false},
		{"10]", false},
		{"", false},
		{"uint256[0]", true},
	}

	for _, test := range tests {
		result := isFixedArray(test.input)
		if result != test.expected {
			t.Errorf("isFixedArray(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}
