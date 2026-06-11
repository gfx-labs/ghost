package abir

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"gfx.cafe/open/ghost/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// Encode writes the Go value v into the [abi.Builder] b using the given
// type descriptors.
//
// With a single type argument, v is encoded directly as that type.
// With multiple type arguments, v must be a struct whose fields correspond
// to the types in order.
//
// Supported Go types for encoding:
//   - uint, uint8..uint64 for unsigned integers
//   - int, int8..int64 for signed integers
//   - uint256.Int and big.Int for 256-bit values
//   - string for address (hex), string, and bytes types
//   - []byte for bytes and bytesN types
//   - bool for bool
//   - structs for tuples
//   - slices for dynamic arrays, arrays for fixed arrays
//
// Panics from the underlying builder are caught and returned as errors.
func Encode(b *abi.Builder, v any, args ...abi.TypeName) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic while encoding: %v", err2)
		}
	}()
	if err != nil {
		return err
	}
	val := reflect.ValueOf(v)
	val = reflect.Indirect(val)
	switch len(args) {
	case 0:
		return fmt.Errorf("Nothing to encode")
	case 1:
		return encode(b, args[0], val)
	default:
		if val.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct type to encode, but got '%v'", val.Kind())
		}
		//fmt.Println(args)
		for i := 0; i < len(args); i++ {
			//fmt.Printf("%s : %v\n", args[i], val.Field(i))
			err = encode(b, args[i], val.Field(i))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// EncodeArray encodes l elements of type t from the slice or array value v.
// Returns an error if v.Len() != l.
func EncodeArray(b *abi.Builder, t abi.TypeName, l int, v reflect.Value) error {
	if l != v.Len() {
		return fmt.Errorf("length mismatch")
	}
	for i := 0; i < l; i++ {
		err := encode(b, t, v.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// t = function signature
// v = golang type containing the values to encode
func encode(b *abi.Builder, t abi.TypeName, v reflect.Value) error {
	st := string(t)
	switch {
	case t.IsTuple():
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct type to encode tuple into, but got '%v'", v.Kind())
		}
		cur := b.EnterTuple()
		targs := t.TupleArgs()
		for i := 0; i < v.NumField(); i++ {
			err := encode(cur, targs[i], v.Field(i))
			if err != nil {
				return err
			}
		}
		cur.Exit()
		return nil
	case t.IsFixedSlice():
		if v.Kind() != reflect.Array {
			return fmt.Errorf("cannot encode %v into %s", v.Kind(), t)
		}
		tn, l := t.UnSlice()
		if l != v.Len() {
			return fmt.Errorf("solidity array length mismatch query: %v target: %v", l, v.Len())
		}
		cur := b.EnterArray(tn, l)
		err := EncodeArray(cur, tn, l, v)
		if err != nil {
			return err
		}
		cur.Exit()
		return nil
	case t.IsSlice():
		if v.Kind() != reflect.Slice {
			return fmt.Errorf("cannot encode %s into %v", t, v.Type())
		}
		tn, _ := t.UnSlice()
		cur := b.EnterDynamicArray()
		err := EncodeArray(cur, tn, v.Len(), v)
		if err != nil {
			return err
		}
		cur.Exit()
		return nil
	case t == abi.ADDRESS:
		err := encodeReflectAddress(b, v)
		return err
	case t == abi.BOOL:
		b.Bool(v.Bool())
		return nil
	case t == abi.STRING:
		b.DString(v.String())
		return nil
	case t == abi.BYTES:
		switch v.Kind() {
		case reflect.String:
			b.DString(v.String())
		case reflect.Slice:
			b.Bytes(v.Bytes())
		default:
			b.Bytes(v.Bytes())
		}
		return nil
	case strings.HasPrefix(st, "fixed"), strings.HasPrefix(st, "ufixed"), strings.HasPrefix(st, "int"), strings.HasPrefix(st, "uint"):
		if st[0] == 'u' {
			if !v.CanUint() {
				ui, ok := v.Interface().(uint256.Int)
				if ok {
					b.BigUint(&ui)
					return nil
				}
				ui2, ok2 := v.Interface().(big.Int)
				if ok2 && ui.Sign() >= 0 {
					b.BigInt(&ui2)
					return nil
				}
				return fmt.Errorf("could not encode %v into %s", v.Kind(), st)
			}
			i := v.Uint()
			b.BigUint(uint256.NewInt(i))
			return nil
		} else {
			//fmt.Println(v.Type())
			if !v.CanInt() {
				//fmt.Println(v.Interface())
				ui, ok := v.Interface().(uint256.Int)
				//fmt.Printf("%v %v\n", ui, ok)
				if ok {
					b.BigUint(&ui)
					return nil
				}
				ui2, ok2 := v.Interface().(big.Int)
				if ok2 {
					b.BigInt(&ui2)
					return nil
				}
				return fmt.Errorf("could not encode %v into %s", v.Kind(), st)
			}
			i := v.Int()
			b.BigInt(big.NewInt(i))
			return nil
		}
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
		return encodeReflectBytes(b, amt, v)
	default:
		return fmt.Errorf("encountered unknown type: %s", st)
	}
}

// n = number of bytes
// will panic if not string or byte slice/array
func encodeReflectBytes(b *abi.Builder, n int, v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		b.FixedBytes(n, []byte(v.String()))
	case reflect.Slice, reflect.Array:
		b.FixedBytes(n, v.Bytes())
	default:
		return fmt.Errorf("could not encode %v into bytes", v.Type())
	}
	return nil
}

func encodeReflectAddress(b *abi.Builder, v reflect.Value) error {
	var addr common.Address
	switch v.Kind() {
	case reflect.String:
		addr = common.HexToAddress(v.String())
	case reflect.Slice, reflect.Array:
		addr = common.BytesToAddress(v.Elem().Bytes())
	default:
		return fmt.Errorf("could not encode %v into %v", v.Type(), addr)
	}
	b.Address(addr)
	return nil
}

// func (b *Builder) NumberReflect(v reflect.Value) error {
// 	ui, err := v.Interface().(uint256.Int)
// 	if !err {
// 		b.BigUint(&ui)
// 		return nil
// 	}
// 	i, err2 := v.Interface().(big.Int)
// 	if !err2 {
// 		b.BigInt(&i)
// 		return nil
// 	}
// 	return fmt.Errorf("could not encode %v into %s", v.Kind(), st)
// }
