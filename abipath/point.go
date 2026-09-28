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
func Point(p string, xs []byte) ([]byte, error) {
	// base is the region offsets are relative to, cur is the read position within it
	base, cur := xs, 0
	for i := 0; i < len(p); {
		switch c := p[i]; {
		case c == '.', c == '/':
			if len(base)-cur < 32 {
				return nil, abi.ErrUnexpectedEOF
			}
			off, ok := abi.WordUint64(base[cur : cur+32])
			if !ok {
				return nil, abi.ErrOffsetOverflow
			}
			if off > uint64(len(base)) {
				return nil, abi.ErrDynamicOverflow
			}
			base, cur = base[off:], 0
			if c == '/' {
				if len(base) < 32 {
					return nil, abi.ErrLenEOF
				}
				base = base[32:]
			}
			i++
		case c >= '0' && c <= '9':
			// n can never exceed the words left in the input, so cap it there to avoid overflow
			limit := (len(base) - cur) / 32
			n := 0
			for ; i < len(p) && p[i] >= '0' && p[i] <= '9'; i++ {
				n = n*10 + int(p[i]-'0')
				if n > limit {
					return nil, ErrOutOfBounds
				}
			}
			cur += n * 32
		default:
			return nil, &InvalidTokenError{Pos: i, Char: c}
		}
	}
	return base[cur:], nil
}

// InvalidTokenError reports an unsupported character in a path.
type InvalidTokenError struct {
	Pos  int
	Char byte
}

func (e *InvalidTokenError) Error() string {
	return fmt.Sprintf("abipath: invalid path token %q at %d", e.Char, e.Pos)
}

func (e *InvalidTokenError) Unwrap() error { return ErrInvalidToken }
