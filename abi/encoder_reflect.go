package abi

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func (b *Builder) Encode(v any, args ...TypeName) (err error) {
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
		return b.encode(args[0], val)
	default:
		if val.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct type to encode, but got '%v'", val.Kind())
		}
		//fmt.Println(args)
		for i := 0; i < len(args); i++ {
			//fmt.Printf("%s : %v\n", args[i], val.Field(i))
			err = b.encode(args[i], val.Field(i))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (b *Builder) EncodeArray(t TypeName, l int, v reflect.Value) error {
	if l != v.Len() {
		return fmt.Errorf("length mismatch")
	}
	for i := 0; i < l; i++ {
		err := b.encode(t, v.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// t = function signature
// v = golang type containing the values to encode
func (b *Builder) encode(t TypeName, v reflect.Value) error {
	st := string(t)
	switch {
	case t.IsTuple():
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct type to encode tuple into, but got '%v'", v.Kind())
		}
		cur := b.EnterTuple()
		targs := t.TupleArgs()
		for i := 0; i < v.NumField(); i++ {
			err := cur.encode(targs[i], v.Field(i))
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
		err := cur.EncodeArray(tn, l, v)
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
		err := cur.EncodeArray(tn, v.Len(), v)
		if err != nil {
			return err
		}
		cur.Exit()
		return nil
	case t == ADDRESS:
		err := b.encodeReflectAddress(v)
		return err
	case t == BOOL:
		b.WriteBool(v.Bool())
		return nil
	case t == STRING:
		b.WriteString(v.String())
		return nil
	case t == BYTES:
		b.WriteBytes(v.String())
		return nil
	case strings.HasPrefix(st, "fixed"), strings.HasPrefix(st, "ufixed"), strings.HasPrefix(st, "int"), strings.HasPrefix(st, "uint"):
		if st[0] == 'u' {
			if !v.CanUint() {
				ui, ok := v.Interface().(uint256.Int)
				if ok {
					b.WriteBigUint(&ui)
					return nil
				}
				ui2, ok2 := v.Interface().(big.Int)
				if ok2 && ui.Sign() >= 0 {
					b.WriteBigInt(&ui2)
					return nil
				}
				return fmt.Errorf("could not encode %v into %s", v.Kind(), st)
			}
			i := v.Uint()
			b.WriteBigUint(uint256.NewInt(i))
			return nil
		} else {
			//fmt.Println(v.Type())
			if !v.CanInt() {
				//fmt.Println(v.Interface())
				ui, ok := v.Interface().(uint256.Int)
				//fmt.Printf("%v %v\n", ui, ok)
				if ok {
					b.WriteBigUint(&ui)
					return nil
				}
				ui2, ok2 := v.Interface().(big.Int)
				if ok2 {
					b.WriteBigInt(&ui2)
					return nil
				}
				return fmt.Errorf("could not encode %v into %s", v.Kind(), st)
			}
			i := v.Int()
			b.WriteBigInt(big.NewInt(i))
			return nil
		}
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
		return b.encodeReflectBytes(amt, v)
	default:
		return fmt.Errorf("encountered unknown type: %s", st)
	}
}

// n = number of bytes
// will panic if not string or byte slice/array
func (b *Builder) encodeReflectBytes(n int, v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		b.WriteFixedBytes(n, []byte(v.String()))
	case reflect.Slice, reflect.Array:
		b.WriteFixedBytes(n, v.Bytes())
	default:
		return fmt.Errorf("could not encode %v into bytes", v.Type())
	}
	return nil
}

func (b *Builder) encodeReflectAddress(v reflect.Value) error {
	var addr common.Address
	switch v.Kind() {
	case reflect.String:
		addr = common.HexToAddress(v.String())
	case reflect.Slice, reflect.Array:
		addr = common.BytesToAddress(v.Elem().Bytes())
	default:
		return fmt.Errorf("could not encode %v into %v", v.Type(), addr)
	}
	b.WriteAddress(addr)
	return nil
}

// func (b *Builder) WriteNumberReflect(v reflect.Value) error {
// 	ui, err := v.Interface().(uint256.Int)
// 	if !err {
// 		b.WriteBigUint(&ui)
// 		return nil
// 	}
// 	i, err2 := v.Interface().(big.Int)
// 	if !err2 {
// 		b.WriteBigInt(&i)
// 		return nil
// 	}
// 	return fmt.Errorf("could not encode %v into %s", v.Kind(), st)
// }
