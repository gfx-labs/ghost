package abi

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func hexDecode(s string) *Decoder {
	s = strings.TrimPrefix(s, "0x")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, `	`, "")
	ans, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewDecoder(ans)
}

func TestThing(t *testing.T) {
	_ = reflect.ValueOf(common.Address{})
}

func TestBasicExample(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000045
0000000000000000000000000000000000000000000000000000000000000001
`)
	type baz struct {
		x int
		y bool
	}

	var r baz
	var err error

	r.x, err = dec.ReadInt()
	if err != nil {
		t.Fatal(err)
	}
	r.y, err = dec.ReadBool()
	if err != nil {
		t.Fatal(err)
	}
	if r.x != 69 {
		t.Errorf("expect %d got %d", 69, r.x)
	}
	if r.y != true {
		t.Errorf("expect %v got %v", true, r.y)
	}
}

func TestBasicExampleType(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000045
0000000000000000000000000000000000000000000000000000000000000001
`)

	One := &big.Int{}
	var Two bool
	//tup := TUPLE(INT32, BOOL)
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

func TestDynamicExampleType(t *testing.T) {

	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000123
0000000000000000000000000000000000000000000000000000000000000080
3132333435363738393000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000456
0000000000000000000000000000000000000000000000000000000000000789
000000000000000000000000000000000000000000000000000000000000000d
48656c6c6f2c20776f726c642100000000000000000000000000000000000000
	`)

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
func TestDynamicAnonymousType(t *testing.T) {

	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000123
0000000000000000000000000000000000000000000000000000000000000080
3132333435363738393000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000456
0000000000000000000000000000000000000000000000000000000000000789
000000000000000000000000000000000000000000000000000000000000000d
48656c6c6f2c20776f726c642100000000000000000000000000000000000000
	`)

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

func TestDynamicExample(t *testing.T) {

	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000123
0000000000000000000000000000000000000000000000000000000000000080
3132333435363738393000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000000000000000e0
0000000000000000000000000000000000000000000000000000000000000002
0000000000000000000000000000000000000000000000000000000000000456
0000000000000000000000000000000000000000000000000000000000000789
000000000000000000000000000000000000000000000000000000000000000d
48656c6c6f2c20776f726c642100000000000000000000000000000000000000
	`)

	type f struct {
		a int
		b []uint32
		c []byte
		d string
	}

	var r f
	var err error
	r.a, err = dec.ReadInt()
	if err != nil {
		t.Fatal("r.a", err)
	}

	arr_b, err := dec.ReadDynamic()
	if err != nil {
		t.Fatal("r.b", err)
	}

	array_len, err := arr_b.ReadInt()
	if err != nil {
		t.Fatal("r.b_len", err)
	}
	r.b = make([]uint32, 0, array_len)
	for i := 0; i < array_len; i++ {
		val, err := arr_b.ReadUint()
		if err != nil {
			t.Fatal("r.b_inside", err)
			t.Fatal(err)
		}
		r.b = append(r.b, uint32(val))
	}

	r.c, err = dec.ReadNPadRight32(10)
	if err != nil {
		t.Fatal(err)
	}
	arr_d, err := dec.ReadDynamic()
	if err != nil {
		t.Fatal("arr_d", err)
	}
	str_len, err := arr_d.ReadInt()
	if err != nil {
		t.Fatal("str_len", err)
	}
	bts, err := arr_d.ReadNPadRight32(str_len)
	if err != nil {
		t.Fatal("str", err)
	}
	r.d = string(bts)
	if !reflect.DeepEqual(r.a, 0x123) {
		t.Errorf("expect %v got %v", 0x123, r.a)
	}
	if !reflect.DeepEqual(r.b, []uint32{0x456, 0x789}) {
		t.Errorf("expect %v got %v", []uint32{0x456, 0x789}, r.b)
	}
	if !(string(r.c) == "1234567890") {
		t.Errorf("expect %v got %v", true, r.c)
	}
}

func TestSimple(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000023
	`)
	type f struct {
		a uint   `abi:"uint256"`
		b []uint `abi:"uint256[]"`
	}
	var r f
	//err := dec.DecodeInto(&r)
	err := dec.Decode(TUPLE(UINT, SLICE(UINT)), &r)
	fmt.Println(r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.a, 7) {
		t.Errorf("expect %v got %v", 7, r.a)
	}
	if !reflect.DeepEqual(r.b, []uint{0x21, 0x22, 0x23}) {
		t.Errorf("expect %v got %v", []uint{0x21, 0x22, 0x23}, r.b)
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
	//err := dec.DecodeInto(&r)
	err := dec.Decode(TUPLE(UINT, SLICE(UINT), ARRAY(BYTES, 2)), &r)
	if err != nil {
		t.Fatal(err)
	}
	if r.a != 7 {
		t.Errorf("expect %v got %v", 7, r.a)
	}
	if !reflect.DeepEqual(r.b, []uint{0x21, 0x22, 0x23}) {
		t.Errorf("expect %v got %v", []uint{0x21, 0x22, 0x23}, r.b)
	}
	if !reflect.DeepEqual(r.c, []string{"abcdefgh", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ"}) {
		t.Errorf("expect %v got %v", []string{"abcdefgh", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ"}, r.c)
	}
}
