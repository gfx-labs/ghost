package abi

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"
)

func TestBasicTypeReflect(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000045
0000000000000000000000000000000000000000000000000000000000000001`)
	One := &big.Int{}
	var Two bool

	err := dec.Decode(INT192, One)
	if err != nil {
		t.Fatal(err)
	}
	err = dec.Decode(BOOL, &Two)
	if err != nil {
		t.Fatal(err)
	}

	if One.Int64() != 69 {
		t.Errorf("expect %d got %d", 69, One)
	}
	if Two != true {
		t.Errorf("expect %v got %v", true, Two)
	}
}

func TestDynamicReflect(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000123
0000000000000000000000000000000000000000000000000000000000000080
3132333435363738393000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000456
0000000000000000000000000000000000000000000000000000000000000789
000000000000000000000000000000000000000000000000000000000000000d
48656c6c6f2c20776f726c642100000000000000000000000000000000000000`)
	type f struct {
		A big.Int `abi:"int64"`
		B []uint  `abi:"uint64[]"`
		C []byte  `abi:"bytes10"`
		D string  `abi:"string"`
	}
	var r f
	err := dec.DecodeInto(&r)
	if err != nil {
		t.Fatal(err)
	}
	if r.A.Int64() != 0x123 {
		t.Errorf("expect %v got %v", 0x123, r.A)
	}
	if !reflect.DeepEqual(r.B, []uint{0x456, 0x789}) {
		t.Errorf("expect %v got %v", []uint{0x456, 0x789}, r.B)
	}
	if !(string(r.C) == "1234567890") {
		t.Errorf("expect %v got %v", "1234567890", string(r.C))
	}
	if r.D != "Hello, world!" {
		t.Errorf("expect %v got %v", "Hello, world!", string(r.D))
	}
}

func TestDynamicTypeNameReflect(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000123
0000000000000000000000000000000000000000000000000000000000000080
3132333435363738393000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000456
0000000000000000000000000000000000000000000000000000000000000789
000000000000000000000000000000000000000000000000000000000000000d
48656c6c6f2c20776f726c642100000000000000000000000000000000000000`)
	type f struct {
		A int
		B []uint
		C []byte
		D string
	}
	var r f
	err := dec.Decode(TUPLE(INT64, SLICE(UINT64), BYTES10, STRING), &r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.A, 0x123) {
		t.Errorf("expect %v got %v", 0x123, r.A)
	}
	if !reflect.DeepEqual(r.B, []uint{0x456, 0x789}) {
		t.Errorf("expect %v got %v", []uint{0x456, 0x789}, r.B)
	}
	if !(string(r.C) == "1234567890") {
		t.Errorf("expect %v got %v", "1234567890", string(r.C))
	}
	if r.D != "Hello, world!" {
		t.Errorf("expect %v got %v", "Hello, world!", string(r.D))
	}
}

func TestSimpleReflect(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000023
	`)
	type f struct {
		A uint   `abi:"uint256"`
		B []uint `abi:"uint256[]"`
	}
	var r f
	err := dec.Decode(TUPLE(UINT, SLICE(UINT)), &r)
	if err != nil {
		t.Fatal(err)
	}
	if r.A != 7 {
		t.Errorf("expect %v got %v", 7, r.A)
	}
	if !reflect.DeepEqual(r.B, []uint{0x21, 0x22, 0x23}) {
		t.Errorf("expect %v got %v", []uint{0x21, 0x22, 0x23}, r.B)
	}
}

func TestComplex(t *testing.T) {
	// 7, 0x60, 7 * 0x20,
	// 		// b
	// 		3, 0x21, 0x22, 0x23,
	// 		// c
	// 		0x40, 0x80,
	// 		8, string("abcdefgh"),
	// 		52, string("ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ")
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000060
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000023
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000080
0000000000000000000000000000000000000000000000000000000000000008
6162636465666768000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000034
4142434445464748494a4b4c4d4e4f505152535455565758595a414243444546
4748494a4b4c4d4e4f505152535455565758595a000000000000000000000000
	`)
	type f struct {
		a uint      `abi:"uint256"`
		b []uint    `abi:"uint256[]"`
		c [2]string `abi:"bytes[2]"`
	}
	var r f
	var err error
	r.a, err = dec.ReadUint()
	if err != nil {
		t.Fatal("r.a", err)
	}
	if r.a != 7 {
		t.Errorf("expect %v got %v", 7, r.a)
	}

	arr_b, err := dec.ReadDynamic()
	if err != nil {
		t.Fatal("r.b", err)
	}

	array_len, err := arr_b.ReadInt()
	if err != nil {
		t.Fatal("r.b_len", err)
	}
	r.b = make([]uint, 0, array_len)
	for i := 0; i < array_len; i++ {
		val, err := arr_b.ReadUint()
		if err != nil {
			t.Fatal("r.b_inside", err)
			t.Fatal(err)
		}
		r.b = append(r.b, uint(val))
	}

	if !reflect.DeepEqual(r.b, []uint{0x21, 0x22, 0x23}) {
		t.Errorf("expect %v got %v", []uint{0x21, 0x22, 0x23}, r.b)
	}

	//err := dec.DecodeInto(&r)
	// 	err := dec.Decode(TUPLE(UINT, SLICE(UINT), ARRAY(BYTES, 2)), &r)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	if r.a != 7 {
	// 		t.Errorf("expect %v got %v", 7, r.a)
	// 	}
	// 	if !reflect.DeepEqual(r.b, []uint{0x21, 0x22, 0x23}) {
	// 		t.Errorf("expect %v got %v", []uint{0x21, 0x22, 0x23}, r.b)
	// 	}
	// 	if !reflect.DeepEqual(r.c, []string{"abcdefgh", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ"}) {
	// 		t.Errorf("expect %v got %v", []string{"abcdefgh", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ"}, r.c)
	// 	}
	// }
}

func TestStructSimple(t *testing.T) {
	dec := hexDecode(`
00000000000000000000000000000000000000000000000000000000000ff010
0000000000000000000000000000000000000000000000000000000000ff0002
6162636400000000000000000000000000000000000000000000000000000000
	`)
	type f struct {
		a int    `abi:"int256"`
		b uint   `abi:"uint256"`
		c []byte `abi:"bytes16"`
	}
	var r f
	var err error
	r.a, err = dec.ReadInt()
	if err != nil {
		t.Fatal("r.a", err)
	}
	if r.a != 0xff010 {
		t.Errorf("expect %v got %v", 0xff010, r.a)
	}
	r.b, err = dec.ReadUint()
	if err != nil {
		t.Fatal("r.b", err)
	}
	if r.b != 0xff0002 {
		t.Errorf("expect %v got %v", 0xff0002, r.b)
	}
	r.c, err = dec.ReadNPadRight32(16)
	r.c = bytes.TrimRight(r.c, "\x00")
	if err != nil {
		t.Fatal(err)
	}
	if !(string(r.c) == "abcd") {
		t.Errorf("expect %v got %v", "abcd", string(r.c))
	}
}

func TestStructComplex(t *testing.T) {

}
