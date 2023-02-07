package abi

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

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
	// switch val.Elem().Kind() {
	// case reflect.Ptr:
	// 	return d.DecodeInto(val.Elem())
	// case reflect.Slice:
	// 	return d.decode(SLICE(BYTES32), val.Elem())
	// case reflect.Struct:
	// 	return d.decode(TUPLE(), val.Elem())
	// default:
	// 	return d.decode(BYTES32, val.Elem())
	// }
}

func CreateTypeName(v any) TypeName {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Slice:
		return SLICE(CreateTypeName(val.Elem()))
	case reflect.Array:
		return ARRAY(CreateTypeName(val.Elem()), val.Len())
	case reflect.Struct:
		var args []TypeName
		for i := 0; i < val.NumField(); i++ {
			tag := val.Type().Field(i).Tag.Get("abi")
			if tag != "" {
				args[i] = TypeName(tag)
			} else {
				args[i] = CreateTypeName(val.Field(i))
			}
		}
		return TUPLE(args...)
	case reflect.Func:
		return FUNCTION
	case reflect.Bool:
		return BOOL
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := strings.ToLower(val.Kind().String())
		return TypeName(s)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := strings.ToLower(val.Kind().String())
		return TypeName(s)
	case reflect.String:
		return STRING
	default:
		return NIL
	}
	return NIL
}

func (d *Decoder) DecodeIntoHelper(v reflect.Value) (err error) {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Slice:
		return d.decode(SLICE(CreateTypeName(val.Elem())), val.Elem())
	case reflect.Array:
		return d.DecodeArray(CreateTypeName(val.Elem()), val.Len(), val.Elem())
	case reflect.Struct:
		var args []TypeName
		for i := 0; i < val.NumField(); i++ {
			args[i] = CreateTypeName(val.Field(i))
		}
		return d.decode(TUPLE(args...), val.Elem())
	case reflect.Func:
		return d.decode(FUNCTION, val)
	case reflect.Bool:
		return d.decode(BOOL, val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// if there is a struct tag, use it

	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	case reflect.String:
		return d.decode(STRING, val)
	default:
		return nil
	}
	return nil
}

func (d *Decoder) Decode(v any, args ...TypeName) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic while decoding: %v", err2)
		}
	}()
	if err != nil {
		return err
	}
	val := reflect.ValueOf(v)
	if !val.CanAddr() || !val.CanSet() {
		return fmt.Errorf("%v cannot be set or addressed", val.Kind())
	}
	val = reflect.Indirect(val)
	switch len(args) {
	case 0:
		return fmt.Errorf("Nothing to decode")
	case 1:
		return d.decode(args[0], val)
	default:
		if val.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct type to args decode into, but got '%v'", val.Kind())
		}
		return d.decode(TUPLE(args...), val)
	}
}

// v is type struct
func (d *Decoder) DecodeTuple(t TypeName, v reflect.Value) (err error) {
	targs := t.TupleArgs()
	//fmt.Println(targs)
	for i := 0; i < v.NumField(); i++ {
		//fmt.Printf("%v: %s\n", i, targs[i])
		//fmt.Printf("before value %v kind %v type %v\n", val.Field(i), val.Field(i).Kind(), val.Field(i).Type())
		err = d.decode(targs[i], v.Field(i))
		if err != nil {
			return err
		}
	}

	return nil
}

// t = type of array
// l = length of array
// v has to be a slice or array or will panic
func (d *Decoder) DecodeArray(t TypeName, l int, v reflect.Value) (err error) {
	vlen := v.Len()
	if l > vlen {
		s := reflect.MakeSlice(v.Type(), l-vlen, l-vlen)
		v.Set(reflect.AppendSlice(v, s))
	}
	for i := 0; i < l; i++ {
		err = d.decode(t, v.Index(i))
		if err != nil {
			return err
		}
	}
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
		fmt.Printf("%s, %v\n", t, target)
		if target.Kind() != reflect.Slice {
			return fmt.Errorf("cannot decode %s into %v", t, target.Type())
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
		fmt.Printf("%s %v %v\n", st, l, target)
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
