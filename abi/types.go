package abi

import (
	"strconv"
	"strings"
)

// represents a valid evm abi type.
type TypeName string

func (t TypeName) IsSlice() bool {
	return strings.HasSuffix(string(t), "[]")
}

func (t TypeName) IsFixedSlice() bool {
	match := regexFixedString(string(t))
	return match
}

func (t TypeName) IsDynamic() bool {
	// tuple check
	dym := false
	if t.IsTuple() {
		for _, arg := range t.TupleArgs() {
			dym = dym || arg.IsDynamic()
		}
		return dym
	}
	// T[k] check
	match := regexDynamic(string(t))
	if match {
		i := strings.Index(string(t), "[")
		t1 := TypeName(t[:i])
		return t1.IsDynamic()
	}
	return t.IsSlice() || t == "bytes" || t == "string"
}

func (t TypeName) IsTuple() bool {
	return strings.HasPrefix(string(t), "(") && strings.HasSuffix(string(t), ")")
}

func (t TypeName) IsSimple() bool {
	return (!t.IsSlice()) && (!t.IsTuple())
}

func (t TypeName) IsNumber() bool {
	st := string(t)
	switch {
	case strings.HasPrefix(st, "fixed"), strings.HasPrefix(st, "ufixed"), strings.HasPrefix(st, "int"), strings.HasPrefix(st, "uint"):
		return true
	}
	return false
}

func (t TypeName) IsUnsigned() bool {
	if len(t) == 0 {
		return false
	}
	return t[0] == 'u'
}

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

func ARRAY(t TypeName, n int) TypeName {
	return TypeName(string(t) + "[" + strconv.Itoa(n) + "]")
}

func SLICE(t TypeName) TypeName {
	return TypeName(t + "[]")
}

func (t TypeName) UnSlice() (TypeName, int) {
	st := string(t)
	i := strings.LastIndexByte(st, '[')
	l, _ := strconv.Atoi(st[i+1 : len(st)-1])

	return TypeName(st[:i]), l
}

func FIXED(M, N int) TypeName {
	return TypeName("fixed" + strconv.Itoa(M) + "x" + strconv.Itoa(N))
}

func UFIXED(M, N int) TypeName {
	return TypeName("ufixed" + strconv.Itoa(M) + "x" + strconv.Itoa(N))
}

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

func pack(elems ...TypeName) TypeName {
	if len(elems) == 1 {
		return elems[0]
	}
	var b strings.Builder
	n := len(elems) - 1
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}
	b.Grow(n)
	b.WriteString(string(elems[0]))
	for _, s := range elems[1:] {
		b.WriteRune(',')
		b.WriteString(string(s))
	}
	return TypeName(b.String())
}

func unpack(t TypeName) []TypeName {
	s := string(t)
	n := len(s) / 6
	if 16 > n {
		n = 16
	}
	out := make([]TypeName, 0, n)
	str := strings.NewReader(s)
	var cur strings.Builder
	state := 0
	for {
		r, _, err := str.ReadRune()
		if err != nil {
			return out
		}
		if state == 0 && r == ')' {
			out = append(out, TypeName(strings.TrimSpace(cur.String())))
			return out
		}
		if r == '(' {
			state = state + 1
		}
		if state > 0 && r == ')' {
			state = state - 1
		}
		if state == 0 && r == ',' {
			out = append(out, TypeName(strings.TrimSpace(cur.String())))
			cur.Reset()
		} else {
			cur.WriteRune(r)
		}
	}
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
