package abi

import (
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000045
0000000000000000000000000000000000000000000000000000000000000001`)
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
	if r.x != 69 {
		t.Errorf("expect %d got %d", 69, r.x)
	}

	r.y, err = dec.ReadBool()
	if err != nil {
		t.Fatal(err)
	}
	if r.y != true {
		t.Errorf("expect %v got %v", true, r.y)
	}
}

func TestSimple(t *testing.T) {
	dec := hexDecode(`
0000000000000000000000000000000000000000000000000000000000000007
0000000000000000000000000000000000000000000000000000000000000040
0000000000000000000000000000000000000000000000000000000000000003
0000000000000000000000000000000000000000000000000000000000000021
0000000000000000000000000000000000000000000000000000000000000022
0000000000000000000000000000000000000000000000000000000000000023`)
	type f struct {
		a uint
		b []uint
	}

	var r f
	var err error
	r.a, err = dec.ReadUint()
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
	r.b = make([]uint, 0, array_len)
	for i := 0; i < array_len; i++ {
		val, err := arr_b.ReadUint()
		if err != nil {
			t.Fatal("r.b_inside", err)
		}
		r.b = append(r.b, uint(val))
	}

	if r.a != 7 {
		t.Errorf("expect %v got %v", 7, r.a)
	}
	if !reflect.DeepEqual(r.b, []uint{0x21, 0x22, 0x23}) {
		t.Errorf("expect %v got %v", []uint{0x21, 0x22, 0x23}, r.b)
	}
}

func TestDynamic(t *testing.T) {
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
