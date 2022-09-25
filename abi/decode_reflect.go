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
			if opts.Contains("topic") {
				continue
			}
			if opts.Contains("arg") {
				continue
			}
			if opts.Contains("method") {
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
