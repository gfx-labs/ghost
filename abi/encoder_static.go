package abi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// Uint256 encodes a 256-bit unsigned integer.
func (d *Builder) Uint256(a *uint256.Int) *Builder {
	d.Word(a.Bytes())
	return d
}

// Address encodes a 20-byte Ethereum address, left-padded to 32 bytes.
func (d *Builder) Address(a common.Address) *Builder {
	d.Word(a[:])
	return d
}

// BigInt encodes a signed integer using two's complement representation.
// Negative values are encoded as their two's complement relative to 2^256.
func (d *Builder) BigInt(ret *big.Int) *Builder {
	nr := new(big.Int).Set(ret)
	if ret.Cmp(new(big.Int)) == -1 {
		nr.Neg(nr)
		nr.Sub(nr, common.Big1)
		nr.Sub(nr, MaxUint256)
		nr.Neg(nr)
		d.Word(nr.Bytes())
		return d
	}
	d.Word(ret.Bytes())
	return d
}

// Bool encodes a boolean as a 32-byte word (0 or 1).
func (d *Builder) Bool(b bool) *Builder {
	if b == true {
		d.Word([]byte{0x1})
		return d
	}
	d.Word([]byte{0x0})
	return d
}

// Int encodes a Go int as a signed ABI integer. Negative values use
// two's complement via [Builder.BigInt].
func (d *Builder) Int(i int) *Builder {
	if i >= 0 {
		return d.Uint(uint(i))
	}
	return d.BigInt(big.NewInt(int64(i)))
}

// Uint encodes a Go uint as a uint256.
func (d *Builder) Uint(i uint) *Builder {
	return d.Uint256(uint256.NewInt(uint64(i)))
}

// Uint8 encodes a uint8, left-padded to 32 bytes.
func (d *Builder) Uint8(i uint8) *Builder {
	return d.Uint(uint(i))
}

// Uint16 encodes a uint16, left-padded to 32 bytes.
func (d *Builder) Uint16(i uint16) *Builder {
	d.Uint(uint(i))
	return d
}

// FixedBytes encodes a fixed-length byte sequence (bytesN where 1 <= N <= 32),
// right-padded to 32 bytes. Panics if len(s) > l.
func (d *Builder) FixedBytes(l int, s []byte) *Builder {
	//fmt.Printf("%v %s %v\n", l, s, len(s))
	if l < len(s) {
		panic("input length mismatch")
	}
	return d.PadRight(s)
}

// Deprecated: BigUint is an alias for [Builder.Uint256]. Use Uint256 instead.
func (d *Builder) BigUint(a *uint256.Int) *Builder {
	return d.Uint256(a)
}
