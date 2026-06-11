package abi

import (
	"strconv"
	"strings"
)

// TypeName represents a Solidity ABI type as a string, such as "uint256",
// "address", "string", "bytes32", "(uint256,address)", or "uint256[]".
//
// Use the package constants (UINT256, ADDRESS, BOOL, etc.) for base types,
// and the constructors [TUPLE], [SLICE], [ARRAY], [FIXED], and [UFIXED]
// for composite types.
type TypeName string

// IsSlice reports whether t is a dynamic array type (ends with "[]").
func (t TypeName) IsSlice() bool {
	return strings.HasSuffix(string(t), "[]")
}

// IsFixedSlice reports whether t is a fixed-size array type (ends with "[N]").
func (t TypeName) IsFixedSlice() bool {
	return isFixedArray(string(t))
}

// IsDynamic reports whether t requires dynamic encoding (offset + data).
// Dynamic types include string, bytes, T[], and any tuple or T[k] containing
// a dynamic element.
func (t TypeName) IsDynamic() bool {
	// tuple check
	dym := false
	if t.IsTuple() {
		for _, arg := range t.TupleArgs() {
			dym = dym || arg.IsDynamic()
		}
		return dym
	}
	if t.IsSlice() {
		return true
	}
	// T[k] check
	match := isFixedArray(string(t))
	if match {
		i := strings.LastIndex(string(t), "[")
		t1 := TypeName(t[:i])
		return t1.IsDynamic()
	}
	return t == "bytes" || t == "string"
}

// IsValid checks if the TypeName represents a valid Ethereum ABI type
func (t TypeName) IsValid() bool {
	// Empty type is invalid
	if t == "" {
		return false
	}

	// Check if it's a tuple
	if t.IsTuple() {
		// Check for proper tuple formatting
		s := string(t)
		if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
			return false
		}

		// Empty tuple "()" is valid
		if s == "()" {
			return true
		}

		// Validate each tuple argument
		args := t.TupleArgs()
		if len(args) == 0 {
			return false
		}
		for _, arg := range args {
			if !arg.IsValid() {
				return false
			}
		}
		return true
	}

	// Check if it's a dynamic array (ends with [])
	if t.IsSlice() {
		s := string(t)
		baseType := TypeName(s[:len(s)-2])
		return baseType.IsValid()
	}

	// Check if it's a fixed array (ends with [n])
	if t.IsFixedSlice() {
		baseType, size := t.UnSlice()
		// Size must be > 0 for valid fixed arrays
		if size <= 0 {
			return false
		}
		return baseType.IsValid()
	}

	// Check base types
	return isValidBaseType(t)
}

// isValidBaseType checks if the type is a valid Ethereum ABI base type
func isValidBaseType(t TypeName) bool {
	s := string(t)

	// Basic types
	switch t {
	case BOOL, ADDRESS, STRING, BYTES:
		return true
	}

	// if it is a tuple or a slice, it's not a base type
	if !t.IsSimple() {
		return false
	}

	// uint types: uint8, uint16, ..., uint256
	if strings.HasPrefix(s, "uint") {
		sizeStr := s[4:]
		if sizeStr == "" {
			return false // "uint" alone is not valid
		}
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return false
		}
		// Must be multiple of 8 and between 8 and 256
		return size >= 8 && size <= 256 && size%8 == 0
	}

	// int types: int8, int16, ..., int256
	if strings.HasPrefix(s, "int") {
		sizeStr := s[3:]
		if sizeStr == "" {
			return false // "int" alone is not valid
		}
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return false
		}
		// Must be multiple of 8 and between 8 and 256
		return size >= 8 && size <= 256 && size%8 == 0
	}

	// bytes types: bytes1, bytes2, ..., bytes32
	if strings.HasPrefix(s, "bytes") && len(s) > 5 {
		sizeStr := s[5:]
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return false
		}
		// Must be between 1 and 32
		return size >= 1 && size <= 32
	}

	// fixed and ufixed types: fixed<M>x<N>, ufixed<M>x<N>
	if strings.HasPrefix(s, "fixed") || strings.HasPrefix(s, "ufixed") {
		prefix := "fixed"
		if strings.HasPrefix(s, "ufixed") {
			prefix = "ufixed"
		}
		rest := s[len(prefix):]

		// Must have format MxN
		cut := strings.Index(rest, "x")
		if cut == -1 {
			return false
		}

		m, err1 := strconv.Atoi(rest[:cut])
		n, err2 := strconv.Atoi(rest[cut+1:])
		if err1 != nil || err2 != nil {
			return false
		}

		// M must be multiple of 8, between 8 and 256
		// N must be between 0 and 80
		return m >= 8 && m <= 256 && m%8 == 0 && n >= 0 && n <= 80
	}

	return false
}

// IsTuple reports whether t is a tuple type (starts with "(" and ends with ")").
func (t TypeName) IsTuple() bool {
	return strings.HasPrefix(string(t), "(") && strings.HasSuffix(string(t), ")")
}

// IsSimple reports whether t is neither a dynamic slice nor a tuple.
// Note that fixed arrays like "uint256[5]" are considered simple.
func (t TypeName) IsSimple() bool {
	return (!t.IsSlice()) && (!t.IsTuple())
}

// IsNumber reports whether t is a numeric type (int, uint, fixed, or ufixed).
func (t TypeName) IsNumber() bool {
	st := string(t)
	switch {
	case strings.HasPrefix(st, "fixed"), strings.HasPrefix(st, "ufixed"), strings.HasPrefix(st, "int"), strings.HasPrefix(st, "uint"):
		return true
	}
	return false
}

// IsUnsigned reports whether t starts with 'u' (uint or ufixed).
func (t TypeName) IsUnsigned() bool {
	if len(t) == 0 {
		return false
	}
	return t[0] == 'u'
}

// TupleArgs parses a tuple type and returns its element types.
// For "(uint256,address)" it returns ["uint256", "address"].
// Handles nested tuples correctly.
func (t TypeName) TupleArgs() []TypeName {
	out := make([]TypeName, 0, 16)
	str := strings.NewReader(string(t))
	var cur strings.Builder
	str.ReadRune()
	state := 0
	for {
		r, _, err := str.ReadRune()
		if err != nil {
			return out
		}
		if state == 0 {
			if r == ')' {
				out = append(out, TypeName(strings.TrimSpace(cur.String())))
				return out
			}
		}
		if r == '(' {
			state = state + 1
		}
		if state > 0 {
			if r == ')' {
				state = state - 1
			}
		}
		if state == 0 && r == ',' {
			out = append(out, TypeName(strings.TrimSpace(cur.String())))
			cur.Reset()
		} else {
			cur.WriteRune(r)
		}
	}
}

// ARRAY constructs a fixed-size array type: ARRAY(UINT256, 5) => "uint256[5]".
func ARRAY(t TypeName, n int) TypeName {
	return TypeName(string(t) + "[" + strconv.Itoa(n) + "]")
}

// SLICE constructs a dynamic array type: SLICE(UINT256) => "uint256[]".
func SLICE(t TypeName) TypeName {
	return TypeName(t + "[]")
}

// UnSlice strips the outermost array suffix and returns the element type
// and length. For dynamic arrays the length is 0:
//
//	SLICE(UINT256).UnSlice()      => ("uint256", 0)
//	ARRAY(UINT256, 5).UnSlice()   => ("uint256", 5)
func (t TypeName) UnSlice() (TypeName, int) {
	st := string(t)
	i := strings.LastIndexByte(st, '[')
	l, _ := strconv.Atoi(st[i+1 : len(st)-1])

	return TypeName(st[:i]), l
}

// FIXED constructs a signed fixed-point type: FIXED(128, 18) => "fixed128x18".
func FIXED(M, N int) TypeName {
	return TypeName("fixed" + strconv.Itoa(M) + "x" + strconv.Itoa(N))
}

// UFIXED constructs an unsigned fixed-point type: UFIXED(128, 18) => "ufixed128x18".
func UFIXED(M, N int) TypeName {
	return TypeName("ufixed" + strconv.Itoa(M) + "x" + strconv.Itoa(N))
}

// TUPLE constructs a tuple type from its elements:
//
//	TUPLE(UINT256, ADDRESS) => "(uint256,address)"
//	TUPLE()                 => "()"
func TUPLE(elems ...TypeName) TypeName {
	switch len(elems) {
	case 0:
		return "()"
	case 1:
		return "(" + elems[0] + ")"
	}
	n := (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.WriteRune('(')
	b.Grow(n)
	b.WriteString(string(elems[0]))
	for _, s := range elems[1:] {
		b.WriteRune(',')
		b.WriteString(string(s))
	}
	b.WriteRune(')')
	return TypeName(b.String())
}

const (
	NIL TypeName = ""

	BOOL     TypeName = "bool"
	FUNCTION TypeName = "function"
	ADDRESS  TypeName = "address"
	STRING   TypeName = "string"
)

const (
	UINT    TypeName = "uint256"
	UINT8   TypeName = "uint8"
	UINT16  TypeName = "uint16"
	UINT24  TypeName = "uint24"
	UINT32  TypeName = "uint32"
	UINT40  TypeName = "uint40"
	UINT48  TypeName = "uint48"
	UINT56  TypeName = "uint56"
	UINT64  TypeName = "uint64"
	UINT72  TypeName = "uint72"
	UINT80  TypeName = "uint80"
	UINT88  TypeName = "uint88"
	UINT96  TypeName = "uint96"
	UINT104 TypeName = "uint104"
	UINT112 TypeName = "uint112"
	UINT120 TypeName = "uint120"
	UINT128 TypeName = "uint128"
	UINT136 TypeName = "uint136"
	UINT144 TypeName = "uint144"
	UINT152 TypeName = "uint152"
	UINT160 TypeName = "uint160"
	UINT168 TypeName = "uint168"
	UINT176 TypeName = "uint176"
	UINT184 TypeName = "uint184"
	UINT192 TypeName = "uint192"
	UINT200 TypeName = "uint200"
	UINT208 TypeName = "uint208"
	UINT216 TypeName = "uint216"
	UINT224 TypeName = "uint224"
	UINT232 TypeName = "uint232"
	UINT240 TypeName = "uint240"
	UINT248 TypeName = "uint248"
	UINT256 TypeName = "uint256"
	INT     TypeName = "int256"
	INT8    TypeName = "int8"
	INT16   TypeName = "int16"
	INT24   TypeName = "int24"
	INT32   TypeName = "int32"
	INT40   TypeName = "int40"
	INT48   TypeName = "int48"
	INT56   TypeName = "int56"
	INT64   TypeName = "int64"
	INT72   TypeName = "int72"
	INT80   TypeName = "int80"
	INT88   TypeName = "int88"
	INT96   TypeName = "int96"
	INT104  TypeName = "int104"
	INT112  TypeName = "int112"
	INT120  TypeName = "int120"
	INT128  TypeName = "int128"
	INT136  TypeName = "int136"
	INT144  TypeName = "int144"
	INT152  TypeName = "int152"
	INT160  TypeName = "int160"
	INT168  TypeName = "int168"
	INT176  TypeName = "int176"
	INT184  TypeName = "int184"
	INT192  TypeName = "int192"
	INT200  TypeName = "int200"
	INT208  TypeName = "int208"
	INT216  TypeName = "int216"
	INT224  TypeName = "int224"
	INT232  TypeName = "int232"
	INT240  TypeName = "int240"
	INT248  TypeName = "int248"
	INT256  TypeName = "int256"
)

const (
	BYTES   TypeName = "bytes"
	BYTES1  TypeName = "bytes1"
	BYTES2  TypeName = "bytes2"
	BYTES3  TypeName = "bytes3"
	BYTES4  TypeName = "bytes4"
	BYTES5  TypeName = "bytes5"
	BYTES6  TypeName = "bytes6"
	BYTES7  TypeName = "bytes7"
	BYTES8  TypeName = "bytes8"
	BYTES9  TypeName = "bytes9"
	BYTES10 TypeName = "bytes10"
	BYTES11 TypeName = "bytes11"
	BYTES12 TypeName = "bytes12"
	BYTES13 TypeName = "bytes13"
	BYTES14 TypeName = "bytes14"
	BYTES15 TypeName = "bytes15"
	BYTES16 TypeName = "bytes16"
	BYTES17 TypeName = "bytes17"
	BYTES18 TypeName = "bytes18"
	BYTES19 TypeName = "bytes19"
	BYTES20 TypeName = "bytes20"
	BYTES21 TypeName = "bytes21"
	BYTES22 TypeName = "bytes22"
	BYTES23 TypeName = "bytes23"
	BYTES24 TypeName = "bytes24"
	BYTES25 TypeName = "bytes25"
	BYTES26 TypeName = "bytes26"
	BYTES27 TypeName = "bytes27"
	BYTES28 TypeName = "bytes28"
	BYTES29 TypeName = "bytes29"
	BYTES30 TypeName = "bytes30"
	BYTES31 TypeName = "bytes31"
	BYTES32 TypeName = "bytes32"
)
