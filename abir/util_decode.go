package abir

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"

	"github.com/gfx-labs/ghost/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type IntSet interface {
	Set(*big.Int) *big.Int
}

type UInt256SetFromBig interface {
	SetFromBig(*big.Int) bool
}
type SetStringErr interface {
	SetString(string, int) error
}

type SetStringOk interface {
	SetString(string, int) bool
}

func reflectBigNumeric(t abi.TypeName, ui *big.Int, target reflect.Value) error {
	switch target.Kind() {
	case reflect.Pointer:
		if target.Type().AssignableTo(typeBigIntPtr) {
			target.Set(reflect.ValueOf(ui))
			return nil
		}
		switch ts := target.Interface().(type) {
		case IntSet:
			ts.Set(ui)
			return nil
		case UInt256SetFromBig:
			ts.SetFromBig(ui)
			return nil
		case SetStringErr:
			ts.SetString(ui.String(), 0)
			return nil
		case SetStringOk:
			ts.SetString(ui.String(), 0)
			return nil
		default:
		}
		t2 := reflect.New(target.Type().Elem())
		target.Set(t2)
		target = t2.Elem()
	default:
	}
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetInt(ui.Int64())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetUint(ui.Uint64())
	case reflect.Float64, reflect.Float32:
		target.SetFloat(float64(ui.Uint64()))
	case reflect.Struct:
		if target.Type().AssignableTo(typeBigInt) {
			target.Set(reflect.ValueOf(*ui))
		} else {
			for fidx := 0; fidx < target.NumField(); fidx++ {
				tag, _ := parseTag(target.Type().Field(fidx).Tag.Get("abi"))
				if tag == "-" {
					continue
				}
				val := target.Field(fidx)
				return reflectBigNumeric(t, ui, val)
			}
		}
	case reflect.Slice:
		switch target.Type().Elem().Kind() {
		case reflect.Uint, reflect.Uint64:
			ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), 4, 4)
			for idx, v := range new(uint256.Int).SetBytes(ui.Bytes()) {
				ns.Index(idx).SetUint(uint64(v))
			}
			target.Set(ns)
		case reflect.Int64, reflect.Int:
			ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), 4, 4)
			for idx, v := range new(uint256.Int).SetBytes(ui.Bytes()) {
				ns.Index(idx).SetInt(int64(v))
			}
			target.Set(ns)
		case reflect.Uint8, reflect.Int8:
			ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), 32, 32)
			for idx, v := range ui.Bytes() {
				ns.Index(idx).SetUint(uint64(v))
			}
			target.Set(ns)
		default:
			return fmt.Errorf("could not slice %v into %v", t, target.Type())
		}
	case reflect.Array:
		switch target.Type().Elem().Kind() {
		case reflect.Uint, reflect.Uint64:
			for idx, v := range new(uint256.Int).SetBytes(ui.Bytes()) {
				target.Index(idx).SetUint(uint64(v))
			}
		case reflect.Int64, reflect.Int:
			for idx, v := range new(uint256.Int).SetBytes(ui.Bytes()) {
				target.Index(idx).SetInt(int64(v))
			}
		case reflect.Uint8, reflect.Int8:
			for idx, v := range ui.Bytes() {
				target.Index(idx).SetUint(uint64(v))
			}
		default:
			return fmt.Errorf("could not array %v into %v", t, target.Type())
		}
	default:
		return fmt.Errorf("could not decode %v into %v", target.Type(), target.Kind())
	}
	return nil
}

func reflectAddress(t abi.TypeName, addr common.Address, target reflect.Value) error {
	switch target.Kind() {
	case reflect.String:
		target.SetString(addr.Hex())
	case reflect.Slice:
		ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), 20, 20)
		for idx, v := range addr {
			ns.Index(idx).SetUint(uint64(v))
		}
		target.Set(ns)
	case reflect.Array:
		if target.Len() < 20 {
			return fmt.Errorf("array too short")
		}
		for idx, v := range addr {
			target.Index(idx).SetUint(uint64(v))
		}
	default:
		return fmt.Errorf("could not decode %v into %v", addr, target.Type())
	}
	return nil
}

func reflectBool(t abi.TypeName, bl bool, target reflect.Value) error {
	switch target.Kind() {
	case reflect.Bool:
		target.SetBool(bl)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if bl {
			target.SetInt(1)
		} else {
			target.SetInt(0)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if bl {
			target.SetUint(1)
		} else {
			target.SetUint(0)
		}
	case reflect.String:
		if bl {
			target.SetString("true")
		} else {
			target.SetString("false")
		}
	}

	return nil
}

func reflectString(t abi.TypeName, str string, target reflect.Value) error {
	switch target.Kind() {
	case reflect.String:
		target.SetString(str)
	case reflect.Slice:
		ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), len(str), len(str))
		target.Set(ns)
		fallthrough
	case reflect.Array:
		if target.Len() < len(str) {
			return fmt.Errorf("array too short")
		}
		for idx, v := range str {
			switch target.Type().Elem().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				target.Index(idx).SetInt(int64(v))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				target.Index(idx).SetUint(uint64(v))
			default:
				return fmt.Errorf("could not decode %v into %v", v, target.Type().Elem().Kind())
			}
		}
	}
	return nil
}

func reflectDynamicBytes(t abi.TypeName, str []byte, target reflect.Value) error {
	target.SetBytes(str)
	return nil
}

func reflectFixedBytes(t abi.TypeName, addr []byte, target reflect.Value) error {
	switch target.Kind() {
	case reflect.String:
		target.SetString(hex.EncodeToString(addr))
	case reflect.Slice:
		ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), len(addr), len(addr))
		for idx, v := range addr {
			ns.Index(idx).SetUint(uint64(v))
		}
		target.Set(ns)
	case reflect.Array:
		if target.Len() < len(addr) {
			return fmt.Errorf("array too short")
		}
		for idx, v := range addr {
			switch target.Type().Elem().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				target.Index(idx).SetInt(int64(v))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				target.Index(idx).SetUint(uint64(v))
			default:
				return fmt.Errorf("could not decode %v into %v", v, target.Type().Elem().Kind())
			}
		}
	default:
		return fmt.Errorf("could not decode %v into %v", addr, target.Type())
	}
	return nil
}
