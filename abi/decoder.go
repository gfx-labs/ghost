package abi

import (
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

var (
	// MaxUint256 is the maximum value that can be represented by a uint256.
	MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
	// MaxInt256 is the maximum value that can be represented by a int256.
	MaxInt256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 255), common.Big1)
)

// Decoder reads ABI-encoded data sequentially from a byte slice.
//
// Each typed method (Uint256, Address, Bool, etc.) reads one or more 32-byte
// words and advances an internal cursor. For dynamic types, use [Decoder.Dynamic]
// or [Decoder.DynamicLength] to follow an offset pointer and obtain a sub-decoder
// positioned at the dynamic data.
//
// The cursor can be inspected with [Decoder.Remaining] and repositioned
// with [Decoder.Seek]. Non-advancing reads are available via [Decoder.Peek],
// [Decoder.PeekWord], and [Decoder.PeekUint256].
type Decoder struct {
	xs  []byte
	cur int
}

func hexDecoder(s string) *Decoder {
	return NewDecoder(hexBytes(s))
}

func hexBytes(s string) []byte {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	ans, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return ans
}

// NewDecoder creates a Decoder reading from xs, starting at position 0.
func NewDecoder(xs []byte) *Decoder {
	return &Decoder{
		xs: xs,
	}
}
// Remaining returns the unread portion of the underlying byte slice.
func (d *Decoder) Remaining() []byte {
	return d.xs[d.cur:]
}

// Peek reads len(o) bytes into o without advancing the cursor.
// Returns [ErrUnexpectedEOF] if insufficient data remains.
func (d *Decoder) Peek(o []byte) (int, error) {
	if (len(d.xs) - d.cur) < len(o) {
		return 0, ErrUnexpectedEOF
	}
	n := copy(o, d.xs[d.cur:d.cur+len(o)])
	return n, nil
}

// PeekWord reads a 32-byte word without advancing the cursor.
func (d *Decoder) PeekWord() (o [32]byte, err error) {
	_, err = d.Peek(o[:])
	if err != nil {
		return
	}
	return o, nil
}

// PeekUint256 reads a uint256 without advancing the cursor.
func (d *Decoder) PeekUint256() (*uint256.Int, error) {
	b, err := d.PeekWord()
	if err != nil {
		return nil, err
	}
	ret := new(uint256.Int).SetBytes(b[:])
	return ret, nil
}

// Read copies len(o) bytes into o and advances the cursor.
// Returns [ErrUnexpectedEOF] if insufficient data remains.
func (d *Decoder) Read(o []byte) (int, error) {
	if (len(d.xs) - d.cur) < len(o) {
		return 0, ErrUnexpectedEOF
	}
	n := copy(o, d.xs[d.cur:d.cur+len(o)])
	d.cur = d.cur + len(o)
	return n, nil
}

// ReadN reads exactly n bytes and advances the cursor.
func (d *Decoder) ReadN(n int) ([]byte, error) {
	o := make([]byte, n)
	_, err := d.Read(o[:])
	if err != nil {
		return nil, err
	}
	return o, nil
}

// ReadWord reads a 32-byte word and advances the cursor.
func (d *Decoder) ReadWord() (o [32]byte, err error) {
	_, err = d.Read(o[:])
	if err != nil {
		return
	}
	return o, nil
}

// ReadNPadRight32 reads n bytes of content and skips the remaining padding
// to align to a 32-byte boundary. Used for decoding bytesN types.
func (d *Decoder) ReadNPadRight32(n int) ([]byte, error) {
	diff := 32 - (n % 32)
	if diff == 32 {
		diff = 0
	}
	o := make([]byte, n)
	_, err := d.Read(o[:])
	if err != nil {
		return nil, err
	}
	_, err = d.ReadN(diff)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Uint256 reads a 32-byte word as an unsigned 256-bit integer.
func (d *Decoder) Uint256() (*uint256.Int, error) {
	b, err := d.ReadWord()
	if err != nil {
		return nil, err
	}
	ret := new(uint256.Int).SetBytes(b[:])
	return ret, nil
}

// Address reads a 32-byte word and extracts the rightmost 20 bytes as an address.
func (d *Decoder) Address() (common.Address, error) {
	word, err := d.ReadWord()
	if err != nil {
		return common.Address{}, err
	}
	ans := common.BytesToAddress(word[:])
	return ans, nil
}

// BigInt reads a 32-byte word as a signed integer using two's complement.
// Values with bit 255 set are interpreted as negative.
func (d *Decoder) BigInt() (*big.Int, error) {
	rt, err := d.Uint256()
	if err != nil {
		return nil, err
	}
	ret := rt.ToBig()
	if ret.Bit(255) == 1 {
		ret.Add(MaxUint256, new(big.Int).Neg(ret))
		ret.Add(ret, common.Big1)
		ret.Neg(ret)
	}
	return ret, nil
}

// Bool reads a 32-byte word and returns false if zero, true otherwise.
func (d *Decoder) Bool() (bool, error) {
	ans, err := d.Uint256()
	if err != nil {
		return false, err
	}
	if ans.Cmp(uint256.NewInt(0)) == 0 {
		return false, nil
	}
	return true, nil
}

// Int reads a signed integer and converts it to a Go int.
// Returns an error if the value overflows int64.
func (d *Decoder) Int() (int, error) {
	ans, err := d.BigInt()
	if err != nil {
		return 0, err
	}
	if !ans.IsInt64() {
		return 0, errors.New("abi: int overflow")
	}
	return int(ans.Int64()), nil
}

// Uint reads an unsigned integer and converts it to a Go uint.
// Returns an error if the value overflows uint64.
func (d *Decoder) Uint() (uint, error) {
	ans, err := d.Uint256()
	if err != nil {
		return 0, err
	}
	if !ans.IsUint64() {
		return 0, errors.New("abi: int overflow")
	}
	return uint(ans.Uint64()), nil
}

// Uint8 reads an unsigned integer and converts it to uint8.
// Returns an error if the value exceeds 255.
func (d *Decoder) Uint8() (uint8, error) {
	ans, err := d.Int()
	if err != nil {
		return 0, err
	}
	if ans > 255 {
		return 0, errors.New("abi: uint8 overflow")
	}
	return uint8(ans), nil
}

// Uint16 reads an unsigned integer and converts it to uint16.
// Returns an error if the value exceeds 65536.
func (d *Decoder) Uint16() (uint16, error) {
	ans, err := d.Int()
	if err != nil {
		return 0, err
	}
	if ans > 65536 {
		return 0, errors.New("abi: uint16 overflow")
	}
	return uint16(ans), nil
}

// Dynamic reads a uint256 offset and returns a new Decoder positioned at
// that byte offset within the original data. Used to follow pointers to
// dynamic data (arrays, strings, bytes, dynamic tuples).
//
// The original decoder's cursor advances past the offset word.
func (d *Decoder) Dynamic() (*Decoder, error) {
	offset, err := d.Uint256()
	if err != nil {
		return nil, err
	}
	actual := int(offset.Uint64())
	if len(d.xs) < actual {
		return nil, errors.New("abi: dynamic overflow")
	}
	return NewDecoder(d.xs[actual:]), nil
}
// DynamicLength reads a dynamic offset, then reads the length prefix at that
// offset. Returns a sub-decoder positioned after the length word and the
// element count. This is the standard pattern for reading dynamic arrays:
//
//	sub, length, err := dec.DynamicLength()
//	for i := 0; i < length; i++ {
//		val, _ := sub.Uint()
//	}
func (d *Decoder) DynamicLength() (*Decoder, int, error) {
	offset, err := d.Uint256()
	if err != nil {
		return nil, 0, err
	}
	actual := int(offset.Uint64())
	if len(d.xs) < actual {
		return nil, 0, errors.New("abi: dynamic overflow")
	}
	// hop over to the new one
	dec1 := NewDecoder(d.xs[actual:])
	l, err := dec1.Int()
	if err != nil {
		return nil, 0, errors.New("abi: len unexpected EOF")
	}
	return NewDecoder(dec1.xs[32:]), l, nil
}

// DString reads a dynamic string by following its offset and length prefix.
func (d *Decoder) DString() (string, error) {
	bts, err := d.Bytes()
	if err != nil {
		return "", err
	}
	return string(bts), nil
}

// Bytes reads dynamic bytes by following the offset and length prefix.
func (d *Decoder) Bytes() ([]byte, error) {
	offset, err := d.Uint256()
	if err != nil {
		return nil, err
	}
	actual := int(offset.Uint64())
	if len(d.xs) < actual {
		return nil, errors.New("abi: dynamic overflow")
	}
	dec := NewDecoder(d.xs[actual:])
	l, err := dec.Uint()
	if err != nil {
		return nil, err
	}
	bts, err := dec.ReadN(int(l))
	if err != nil {
		return nil, err
	}
	return bts, nil
}

// Seek repositions the cursor using io.SeekStart, io.SeekCurrent, or
// io.SeekEnd semantics. Returns the new absolute position.
func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	startByte := d.cur
	switch whence {
	case io.SeekStart:
		startByte = int(offset)
	case io.SeekCurrent:
		startByte = d.cur + int(offset)
	case io.SeekEnd:
		startByte = len(d.xs) - int(offset)
	}
	if startByte < 0 || startByte > len(d.xs) {
		return int64(startByte), errors.New("invalid seek offset")
	}
	d.cur = startByte
	return int64(startByte), nil
}
