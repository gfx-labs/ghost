package aq

import (
	"bufio"
	"log"
	"math/big"
	"strings"

	"gfx.cafe/open/ghost/abi"
)

type Symbol string

type View struct {
	b []byte
}

func NewView(b []byte) *View {
	return &View{
		b: b,
	}
}

func (v *View) Search(str string) string {
	tf := []TransformFunc{}
	rd := bufio.NewReader(strings.NewReader(str))
	for {
		v, _, err := rd.ReadRune()
		if err != nil {
			break
		}
		switch v {
		case '.':
			tf = append(tf, Identity)
		case '(':
			args, err := rd.ReadString(')')
			if err == nil {
				args = args[:len(args)-1]
			}
			tf = append(tf, TupleIndex(args))
		case '{':
			args, err := rd.ReadString('}')
			if err == nil {
				args = args[:len(args)-1]
			}
			tf = append(tf, TupleHop(args))
		case '[':
			args, err := rd.ReadString(']')
			if err == nil {
				args = args[:len(args)-1]
			}
			tf = append(tf, ArrayIndex(args))
		case '<':
			args, err := rd.ReadString('>')
			if err == nil {
				args = args[:len(args)-1]
			}
			tf = append(tf, ByteArrayIndex(args))
		default:
			break
		}
	}
	o := [][]byte{v.b}
	for _, v := range tf {
		this := [][]byte{}
		for _, co := range o {
			this = append(this, v(co)...)
		}
		o = this
	}
	if len(o) == 1 {
		return abi.PrettyHex(o[0])
	}
	sb := new(strings.Builder)
	sb.WriteRune('[')
	for idx, v := range o {
		if idx != 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(abi.PrettyHex(v))
	}
	sb.WriteRune(']')
	return sb.String()
}

type TransformFunc = func([]byte) [][]byte

func Identity(xs []byte) [][]byte {
	return [][]byte{xs}
}

func TupleHop(args ...string) func(xs []byte) [][]byte {
	if len(args) == 1 && args[0] == "" {
		args = nil
	}
	return func(xs []byte) [][]byte {
		var offset int
		dec := abi.NewDecoder(xs)
		for i := 0; i <= atoi(args[0]); i++ {
			offset, _ = dec.ReadInt()
		}
		if offset > len(xs) {
			return nil
		}
		return [][]byte{xs[offset:]}
	}
}

func Hop(args ...string) func(xs []byte) [][]byte {
	if len(args) == 1 && args[0] == "" {
		args = nil
	}
	return func(xs []byte) [][]byte {
		o := [][]byte{}
		for _, v := range args {
			idx := atoi(v)
			offset := idx
			if offset > len(xs) {
				return nil
			}
			o = append(o, xs[offset:])
		}
		return o
	}
}

func ByteArrayIndex(args ...string) func(xs []byte) [][]byte {
	if len(args) == 1 && args[0] == "" {
		args = nil
	}
	return func(xs []byte) [][]byte {
		log.Println(abi.PrettyHex(xs))
		dec := abi.NewDecoder(xs)
		ln, _ := dec.ReadInt()
		o := [][]byte{}
		if len(args) == 0 {
			for idx := 0; idx < ln; idx++ {
				str, err := dec.ReadString()
				log.Println(str, err)
				if err != nil {
					continue
				}
				o = append(o, []byte(str))
			}
		} else {
			for range args {
				// todo: implemnet
			}
		}
		return o
	}
}

func ArrayIndex(args ...string) func(xs []byte) [][]byte {
	if len(args) == 1 && args[0] == "" {
		args = nil
	}
	return func(xs []byte) [][]byte {
		dec := abi.NewDecoder(xs)
		ln, _ := dec.ReadInt()
		o := [][]byte{}
		if len(args) == 0 {
			for idx := 0; idx < ln; idx++ {
				offset := 32*idx + 32
				if offset > len(xs) {
					continue
				}
				o = append(o, xs[offset:])
			}
		} else {
			for _, v := range args {
				idx := atoi(v)
				offset := 32*idx + 32
				if offset > len(xs) {
					continue
				}
				o = append(o, xs[offset:])
			}
		}
		return o
	}
}
func TupleIndex(args ...string) func(xs []byte) [][]byte {
	if len(args) == 1 && args[0] == "" {
		args = nil
	}
	return func(xs []byte) [][]byte {
		o := [][]byte{}
		for _, v := range args {
			idx := atoi(v)
			offset := 32 * idx
			if offset+32 > len(xs) {
				return nil
			}
			o = append(o, xs[offset:])
		}
		return o
	}
}

func atoi(s string) int {
	i, _ := new(big.Int).SetString(s, 0)
	if i == nil {
		return 0
	}
	return int(i.Int64())
}
