package abi

// isFixedArray checks if string ends with a fixed array pattern [n] where n is a number
func isFixedArray(s string) bool {
	n := len(s)
	if n < 3 { // minimum is "[0]"
		return false
	}

	// Check if it ends with ']'
	if s[n-1] != ']' {
		return false
	}

	// Find the matching '[' going backwards
	bracketPos := -1
	for i := n - 2; i >= 0; i-- {
		if s[i] == '[' {
			bracketPos = i
			break
		}
		// Must be a digit
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	// Must have found '[' and have at least one digit between '[' and ']'
	return bracketPos >= 0 && bracketPos < n-2
}
