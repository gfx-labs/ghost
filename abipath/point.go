package abipath

import (
	"errors"
	"fmt"

	"github.com/gfx-labs/ghost/abi"
)

var (
	ErrOutOfBounds  = fmt.Errorf("abipath: index out of bounds: %w", abi.ErrUnexpectedEOF)
	ErrInvalidToken = errors.New("abipath: invalid path token")
)

// Point walks xs according to path p and returns the remaining bytes.
//
//	"."  follow a dynamic offset
//	"/"  follow a dynamic offset and skip its length word
//	"N"  skip N 32-byte words
//
// The returned slice aliases xs.
func Point(p string, xs []byte) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			out, err = nil, fmt.Errorf("abipath: %v", r)
		}
	}()
	dec := abi.NewDecoder(xs)
	for i := 0; i < len(p); {
		switch c := p[i]; {
		case c == '.':
			if dec, err = dec.Dynamic(); err != nil {
				return nil, err
			}
			i++
		case c == '/':
			if dec, _, err = dec.DynamicLength(); err != nil {
				return nil, err
			}
			i++
		case c >= '0' && c <= '9':
			// n can never exceed the word count of the input, so cap it there to avoid overflow
			limit := len(dec.Remaining()) / 32
			n := 0
			for ; i < len(p) && p[i] >= '0' && p[i] <= '9'; i++ {
				n = n*10 + int(p[i]-'0')
				if n > limit {
					return nil, ErrOutOfBounds
				}
			}
			if err = dec.Skip(n * 32); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("%w: %q at %d", ErrInvalidToken, c, i)
		}
	}
	return dec.Remaining(), nil
}
