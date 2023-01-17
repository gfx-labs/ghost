package abi

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

type IntSet interface {
	Set(*big.Int) *big.Int
}
type SetStringErr interface {
	SetString(string, int) error
}

type SetStringOk interface {
	SetString(string, int) bool
}

// TODO: DONT USE THIS
func (d *Decoder) DecodeInto(v any) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic while decoding: %v", err2)
		}
	}()
	if err != nil {
		return err
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("abi: expected ptr type to decode into, but got '%v'", val.Kind())
	}
	switch val.Elem().Kind() {
	case reflect.Ptr:
		return d.DecodeInto(val.Elem())
	case reflect.Slice:
		return d.decode(SLICE(BYTES32), val.Elem())
	case reflect.Struct:
		return d.decode(TUPLE(), val.Elem())
	default:
		return d.decode(BYTES32, val.Elem())
	}
}

// wrapper function
func (d *Decoder) Decode(t TypeName, v any) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic while decoding: %v", err2)
		}
	}()
	if err != nil {
		return err
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("abi: expected ptr type to decode into, but got '%v'", val.Kind())
	}
	return d.decode(t, val.Elem())
}

// st is a struct
func (d *Decoder) DecodeTuple(t TypeName, val reflect.Value) (err error) {
	targs := t.TupleArgs()
	for i := 0; i < val.NumField(); i++ {
		err = d.decode(targs[i], reflect.ValueOf(val.Type().Field(i)))
		if err != nil {
			return err
		}
	}

	return nil
}

// t = type of array
// l = length of array
// tgt has to be a slice or array
func (d *Decoder) DecodeArray(t TypeName, l int, target reflect.Value) (err error) {
	ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), l, l)
	for i := 0; i < l; i++ {
		err = d.decode(t, ns.Index(i))
		if err != nil {
			return err
		}
	}
	target.Set(ns)
	return nil
}

// TODO:
// implement map, which should populate m[0], m[1]... etc
// decode takes in a reflect.Value that points to the actual thing.
func (dec *Decoder) decode(t TypeName, target reflect.Value) error {
	st := (string)(t)
	switch {
	case t.IsTuple():
		if target.Kind() != reflect.Struct {
			return fmt.Errorf("abi: expected struct type to decode tuple into, but got '%v'", target.Kind())
		}
		if t.IsDynamic() {
			// read dynamic offset
			cur, err2 := dec.ReadDynamic()
			if err2 != nil {
				return err2
			}
			return cur.DecodeTuple(t, target)
		}
		return dec.DecodeTuple(t, target)
	case t.IsSlice():
		if target.Kind() != reflect.Slice {
			return fmt.Errorf("cannot decode %s into %v", t, target.Kind())
		}
		// read dynamic offset
		cur, err2 := dec.ReadDynamic()
		if err2 != nil {
			return err2
		}
		// read the length
		l, err2 := cur.ReadInt()
		if err2 != nil {
			return err2
		}
		st, _ := t.UnSlice()
		return cur.DecodeArray(st, l, target)
	case t.IsFixedSlice():
		if target.Kind() != reflect.Array {
			return fmt.Errorf("cannot decode %s into %v", t, target.Kind())
		}
		// check type, if dynamic or not
		tn, l := t.UnSlice()
		if l != target.Len() {
			return fmt.Errorf("solidity array length mismatch query: %v target: %v", l, target.Len())
		}
		if tn.IsDynamic() {
			cur, err2 := dec.ReadDynamic()
			if err2 != nil {
				return err2
			}
			return cur.DecodeArray(tn, l, target)
		}
		return dec.DecodeArray(tn, l, target)
	case strings.HasPrefix(st, "fixed"), strings.HasPrefix(st, "ufixed"), strings.HasPrefix(st, "int"), strings.HasPrefix(st, "uint"):
		var ui *big.Int
		var err error
		if st[0] == 'u' {
			uib, err := dec.ReadBigUint()
			if err != nil {
				return err
			}
			ui = uib.ToBig()
		} else {
			ui, err = dec.ReadBigInt()
			if err != nil {
				return err
			}
		}
		err = reflectBigNumeric(t, ui, target)
		if err != nil {
			return err
		}
		return nil
	case t == ADDRESS:
		addr, err := dec.ReadAddress()
		if err != nil {
			return err
		}
		return reflectAddress(t, addr, target)
	case t == BOOL:
		bl, err := dec.ReadBool()
		if err != nil {
			return err
		}
		return reflectBool(t, bl, target)
	case t == STRING:
		str, err := dec.ReadString()
		if err != nil {
			return err
		}
		return reflectString(t, str, target)
	case t == BYTES:
		sub, err := dec.ReadDynamic()
		if err != nil {
			return err
		}
		l, err := sub.ReadInt()
		if err != nil {
			return err
		}
		bts, err := sub.ReadN(l)
		if err != nil {
			return err
		}
		return reflectDynamicBytes(t, bts, target)
	case strings.HasPrefix(st, "bytes") || t == FUNCTION:
		var amt int
		var err error
		if t == FUNCTION {
			amt = 24
		} else {
			bts := strings.TrimPrefix(st, "bytes")
			amt, err = strconv.Atoi(bts)
			if err != nil {
				return err
			}
		}
		bts, err := dec.ReadNPadRight32(amt)
		if err != nil {
			return err
		}
		return reflectFixedBytes(t, bts, target)
	default:
		return fmt.Errorf("encountered unknown type: %s", st)
	}
}
