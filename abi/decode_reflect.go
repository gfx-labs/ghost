package abi

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/holiman/uint256"
)

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
	//TODO: smartly pick the type to decode to based on the value of v
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

func (d *Decoder) decodeTuple(t TypeName, st reflect.Value) (err error) {
	ty := st.Type()
	switch st.Kind() {
	case reflect.Slice:
		for idx, t := range t.TupleArgs() {
			err := d.decode(t, st.Index(idx))
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		args := t.TupleArgs()
		argidx := 0
		for i := 0; i < st.NumField(); i++ {
			fld := ty.Field(i)
			if fld.Anonymous {
				t := fld.Type
				if t.Kind() == reflect.Pointer {
					t = t.Elem()
				}
				if !fld.IsExported() && t.Kind() != reflect.Struct {
					continue
				}
			} else if !fld.IsExported() {
				continue
			}
			tag := fld.Tag.Get("abi")
			if tag == "-" {
				continue
			}
			name, opts := parseTag(tag)
			if !isValidTag(name) {
				name = ""
			}
			if opts.Contains("-") {
				continue
			}
			if opts.Contains("input") {
				continue
			}
			if name == "" {
				if argidx < len(args) {
					name = string(args[argidx])
				}
			}
			argidx = argidx + 1
			tfld := st.FieldByName(fld.Name)
			switch tfld.Kind() {
			}
			err := d.decode(TypeName(name), tfld)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("cannot unmarshal tuple into %v", st.Kind())
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
		return dec.decodeTuple(t, target)
	case t.IsSlice():
		// is solidity array
		switch target.Kind() {
		case reflect.Slice:
		default:
			return fmt.Errorf("cannot decode %s into %v", t, target.Kind())
		}
		// read dynamic offset
		cur, err2 := dec.ReadDynamic()
		if err2 != nil {
			return err2
		}
		// read the amount
		leng, err2 := cur.ReadInt()
		if err2 != nil {
			return err2
		}
		ur := t.UnSlice()
		ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), leng, leng)
		for i := 0; i < leng; i++ {
			cur.decode(ur, ns.Index(i))
		}
		target.Set(ns)
		return nil
	case strings.HasPrefix(st, "fixed"), strings.HasPrefix(st, "ufixed"), strings.HasPrefix(st, "int"), strings.HasPrefix(st, "uint"):
		var ui *uint256.Int
		var err error
		if st[0] == 'u' {
			ui, err = dec.ReadBigUint()
			if err != nil {
				return err
			}
		} else {
			ui, err = dec.ReadBigInt()
			if err != nil {
				return err
			}
		}
		switch target.Kind() {
		case reflect.Pointer:
			if target.Type().AssignableTo(typeBigIntPtr) {
				target.Set(reflect.ValueOf(*ui.ToBig()))
				return nil
			}
			t2 := reflect.New(target.Type().Elem())
			target.Set(t2)
			target = t2.Elem()
		default:
		}
		switch target.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target.SetInt(ui.ToBig().Int64())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			target.SetUint(ui.Uint64())
		case reflect.Float64, reflect.Float32:
			target.SetFloat(float64(ui.Uint64()))
		case reflect.Struct:
			if target.Type().AssignableTo(typeBigInt) {
				target.Set(reflect.ValueOf(ui.ToBig()))
			}
		case reflect.Slice:
			switch target.Type().Elem().Kind() {
			case reflect.Uint, reflect.Int, reflect.Uint64, reflect.Int64:
				ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), 4, 4)
				for idx, v := range ui {
					ns.Index(idx).SetUint(uint64(v))
				}
				target.Set(ns)
			case reflect.Uint8, reflect.Int8:
				ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), 32, 32)
				for idx, v := range ui.Bytes() {
					ns.Index(idx).SetUint(uint64(v))
				}
				target.Set(ns)
			default:
				return fmt.Errorf("could not slice %v into %v", st, target.Type())
			}
		case reflect.Array:
			switch target.Type().Elem().Kind() {
			case reflect.Uint, reflect.Int, reflect.Uint64, reflect.Int64:
				for idx, v := range ui {
					target.Index(idx).SetUint(uint64(v))
				}
			case reflect.Uint8, reflect.Int8:
				for idx, v := range ui.Bytes32() {
					target.Index(idx).SetUint(uint64(v))
				}
			default:
				return fmt.Errorf("could not array %v into %v", st, target.Type())
			}
		default:
			return fmt.Errorf("could not decode %v into %v", target.Type(), target.Kind())
		}
		return nil
	case t == ADDRESS:
		addr, err := dec.ReadAddress()
		if err != nil {
			return err
		}
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
	case t == BOOL:
		bl, err := dec.ReadBool()
		if err != nil {
			return err
		}
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
			return fmt.Errorf("could not decode %v into %v", bl, target.Type())
		}
		return nil
	case t == STRING:
		str, err := dec.ReadString()
		if err != nil {
			return err
		}
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
		target.SetBytes(bts)
		return nil
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
		addr, err := dec.ReadNPadRight32(amt)
		if err != nil {
			return err
		}
		switch target.Kind() {
		case reflect.String:
			target.SetString(hex.EncodeToString(addr))
		case reflect.Slice:
			ns := reflect.MakeSlice(reflect.SliceOf(target.Type().Elem()), amt, amt)
			for idx, v := range addr {
				ns.Index(idx).SetUint(uint64(v))
			}
			target.Set(ns)
		case reflect.Array:
			if target.Len() < amt {
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
	default:
		return fmt.Errorf("encountered unknown type: %s", st)
	}
}
