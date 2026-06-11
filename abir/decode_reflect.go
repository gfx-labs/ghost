package abir

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/gfx-labs/ghost/abi"
)

// DecodeBytes is a convenience that creates a [abi.Decoder] from raw bytes
// and calls [Decode].
func DecodeBytes(xs []byte, v any, hint ...abi.TypeName) (err error) {
	return Decode(abi.NewDecoder(xs), v, hint...)
}

// DecodeInto decodes ABI data into v by inferring ABI types from the Go type
// of v using [CreateTypeName]. The value v must be a pointer. For struct types,
// the struct fields are decoded as tuple arguments; for scalar types, a single
// value is decoded.
func DecodeInto(d *abi.Decoder, v any) (err error) {
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

	val = val.Elem()
	tn := CreateTypeName(val.Type())
	switch val.Kind() {
	case reflect.Struct, reflect.Pointer:
		return Decode(d, v, tn.TupleArgs()...)
	}
	return Decode(d, v, tn)
}

// CreateTypeName maps a Go reflect.Type to its corresponding [abi.TypeName].
// Known types (uint256.Int, big.Int, common.Address, common.Hash) are mapped
// to their canonical ABI names. Struct fields with `abi` tags override the
// inferred type. Fields tagged `abi:"-"` are skipped.
func CreateTypeName(t reflect.Type) abi.TypeName {
	switch t {
	case typeUint256, typeUint256Ptr, typeBigInt, typeBigIntPtr:
		return abi.UINT256
	case typeCommonHash, typeCommonHashPtr:
		return abi.BYTES32
	case typeCommonAddress, typeCommonAddressPtr:
		return abi.ADDRESS
	default:
	}
	switch t.Kind() {
	case reflect.Pointer:
		return CreateTypeName(t.Elem())
	case reflect.Slice:
		return abi.SLICE(CreateTypeName(t.Elem()))
	case reflect.Array:
		// special bytes type
		if t.Elem().Kind() == reflect.Uint8 && t.Len() <= 32 {
			return abi.TypeName(string(abi.BYTES) + strconv.Itoa(t.Len()))
		}
		return abi.ARRAY(CreateTypeName(t.Elem()), t.Len())
	case reflect.Struct:
		args := make([]abi.TypeName, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			tag, _ := parseTag(t.Field(i).Tag.Get("abi"))
			if tag == "-" {
				continue
			} else if tag != "" {
				args = append(args, abi.TypeName(tag))
			} else {
				args = append(args, CreateTypeName(t.Field(i).Type))
			}
		}
		return abi.TUPLE(args...)
	case reflect.Func:
		return abi.FUNCTION
	case reflect.Bool:
		return abi.BOOL
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := strings.ToLower(t.Kind().String())
		return abi.TypeName(s)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := strings.ToLower(t.Kind().String())
		return abi.TypeName(s)
	case reflect.String:
		return abi.STRING
	default:
		return abi.NIL
	}
}

// Decode reads ABI-encoded data from d into v using the given type descriptors.
//
// With a single type argument, v can be a pointer to any supported Go type.
// With multiple type arguments, v must be a pointer to a struct whose fields
// correspond to the types in order. Fields tagged `abi:"-"` are skipped.
//
// Panics from the underlying decoder are caught and returned as errors.
func Decode(d *abi.Decoder, v any, args ...abi.TypeName) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic while decoding: %v", err2)
		}
	}()
	if err != nil {
		return err
	}
	val := reflect.ValueOf(v)
	val = reflect.Indirect(val)
	if !val.CanAddr() || !val.CanSet() {
		return fmt.Errorf("%v cannot be set or addressed", val.Kind())
	}
	switch len(args) {
	case 0:
		return fmt.Errorf("Nothing to decode")
	case 1:
		// an assumption is made here that if the length of args is 1, then we can directly
		// decode into v
		return decode(d, args[0], val)
	default:
		if val.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct type to args decode into, but got '%v'", val.Kind())
		}
		//return d.decode(TUPLE(args...), val)
		fidx := 0
		for i := 0; i < len(args); i++ {
			//fmt.Printf("%v -> %v\n", args[i], val.Field(i))
		skiptag:
			tag, _ := parseTag(val.Type().Field(fidx).Tag.Get("abi"))
			if tag == "-" {
				fidx = fidx + 1
				goto skiptag
			}
			err = decode(d, args[i], val.Field(fidx))
			if err != nil {
				return err
			}
			fidx++
		}
		return nil
	}
}

// DecodeTuple decodes a tuple type t into the struct value v.
// Each tuple element is decoded into the corresponding struct field.
func DecodeTuple(d *abi.Decoder, t abi.TypeName, v reflect.Value) (err error) {
	targs := t.TupleArgs()
	//fmt.Println(targs)
	for i := 0; i < v.NumField(); i++ {
		//fmt.Printf("%v: %s\n", i, targs[i])
		//fmt.Printf("before value %v kind %v type %v\n", val.Field(i), val.Field(i).Kind(), val.Field(i).Type())
		err = decode(d, targs[i], v.Field(i))
		if err != nil {
			return err
		}
	}

	return nil
}

// DecodeArray decodes l elements of type t into the slice or array value v.
// If v is a slice shorter than l, it is grown to fit.
func DecodeArray(d *abi.Decoder, t abi.TypeName, l int, v reflect.Value) (err error) {
	vlen := v.Len()
	if l > vlen {
		s := reflect.MakeSlice(v.Type(), l-vlen, l-vlen)
		v.Set(reflect.AppendSlice(v, s))
	}
	for i := 0; i < l; i++ {
		err = decode(d, t, v.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO:
// implement map, which should populate m[0], m[1]... etc
// decode takes in a reflect.Value that points to the actual thing.
func decode(dec *abi.Decoder, t abi.TypeName, target reflect.Value) error {
	st := string(t)
	switch {
	case t.IsTuple():
		if target.Kind() != reflect.Struct {
			return fmt.Errorf("abi: expected struct type to decode tuple into, but got '%v'", target.Kind())
		}
		if t.IsDynamic() {
			// read dynamic offset
			cur, err2 := dec.Dynamic()
			if err2 != nil {
				return err2
			}
			return DecodeTuple(cur, t, target)
		}
		return DecodeTuple(dec, t, target)
	case t.IsSlice():
		if target.Kind() != reflect.Slice {
			return fmt.Errorf("cannot decode %s into %v", t, target.Type())
		}
		// read dynamic offset
		cur, l, err2 := dec.DynamicLength()
		if err2 != nil {
			return err2
		}
		st, _ := t.UnSlice()
		return DecodeArray(cur, st, l, target)
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
			cur, err2 := dec.Dynamic()
			if err2 != nil {
				return err2
			}
			return DecodeArray(cur, tn, l, target)
		}
		return DecodeArray(dec, tn, l, target)
	case strings.HasPrefix(st, "fixed"), strings.HasPrefix(st, "ufixed"), strings.HasPrefix(st, "int"), strings.HasPrefix(st, "uint"):
		var ui *big.Int
		var err error
		if st[0] == 'u' {
			uib, err := dec.Uint256()
			if err != nil {
				return err
			}
			ui = uib.ToBig()
		} else {
			ui, err = dec.BigInt()
			if err != nil {
				return err
			}
		}
		err = reflectBigNumeric(t, ui, target)
		if err != nil {
			return err
		}
		return nil
	case t == abi.ADDRESS:
		addr, err := dec.Address()
		if err != nil {
			return err
		}
		return reflectAddress(t, addr, target)
	case t == abi.BOOL:
		bl, err := dec.Bool()
		if err != nil {
			return err
		}
		return reflectBool(t, bl, target)
	case t == abi.STRING, t == abi.BYTES:
		str, err := dec.DString()
		if err != nil {
			return err
		}
		if t == abi.BYTES && target.Kind() == reflect.Slice && target.Type().Elem().Kind() == reflect.Uint8 {
			return reflectDynamicBytes(t, []byte(str), target)
		}
		return reflectString(t, str, target)
	case strings.HasPrefix(st, "bytes") || t == abi.FUNCTION:
		var amt int
		var err error
		if t == abi.FUNCTION {
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
