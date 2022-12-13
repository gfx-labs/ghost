package abi

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

var (
	// MaxUint256 is the maximum value that can be represented by a uint256.
	MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
	// MaxInt256 is the maximum value that can be represented by a int256.
	MaxInt256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 255), common.Big1)
)

type Decoder struct {
	xs  []byte
	cur int
}

func NewDecoder(xs []byte) *Decoder {
	return &Decoder{
		xs: xs,
	}
}

func (d *Decoder) ReadWord() (o [32]byte, err error) {
	_, err = d.Read(o[:])
	if err != nil {
		return
	}
	return o, nil
}

func (d *Decoder) ReadN(n int) ([]byte, error) {
	o := make([]byte, n)
	_, err := d.Read(o[:])
	if err != nil {
		return nil, err
	}
	return o, nil
}

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

func (d *Decoder) Read(o []byte) (int, error) {
	if (len(d.xs) - d.cur) < len(o) {
		return 0, errors.New("abi: unexpected EOF")
	}
	n := copy(o, d.xs[d.cur:d.cur+len(o)])
	d.cur = d.cur + len(o)
	return n, nil
}

func (d *Decoder) ReadBigUint() (*uint256.Int, error) {
	b, err := d.ReadWord()
	if err != nil {
		return nil, err
	}
	ret := new(uint256.Int).SetBytes(b[:])
	return ret, nil
}

func (d *Decoder) ReadAddress() (common.Address, error) {
	word, err := d.ReadWord()
	if err != nil {
		return common.Address{}, err
	}
	ans := common.BytesToAddress(word[:])
	return ans, nil
}

func (d *Decoder) ReadBigInt() (*big.Int, error) {
	rt, err := d.ReadBigUint()
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

func (d *Decoder) ReadBool() (bool, error) {
	ans, err := d.ReadBigUint()
	if err != nil {
		return false, err
	}
	if ans.Cmp(uint256.NewInt(0)) == 0 {
		return false, nil
	}
	return true, nil
}

func (d *Decoder) ReadInt() (int, error) {
	ans, err := d.ReadBigInt()
	if err != nil {
		return 0, err
	}
	if !ans.IsInt64() {
		return 0, errors.New("abi: int overflow")
	}
	return int(ans.Int64()), nil
}

func (d *Decoder) ReadUint() (uint, error) {
	ans, err := d.ReadBigUint()
	if err != nil {
		return 0, err
	}
	if !ans.IsUint64() {
		return 0, errors.New("abi: int overflow")
	}
	return uint(ans.Uint64()), nil
}

func (d *Decoder) ReadUint8() (uint8, error) {
	ans, err := d.ReadInt()
	if err != nil {
		return 0, err
	}
	if ans > 255 {
		return 0, errors.New("abi: uint8 overflow")
	}
	return uint8(ans), nil
}

func (d *Decoder) ReadUint16() (uint16, error) {
	ans, err := d.ReadInt()
	if err != nil {
		return 0, err
	}
	if ans > 65536 {
		return 0, errors.New("abi: uint16 overflow")
	}
	return uint16(ans), nil
}

func (d *Decoder) ReadDynamic() (*Decoder, error) {
	offset, err := d.ReadBigUint()
	if err != nil {
		return nil, err
	}
	actual := int(offset.Uint64())
	if len(d.xs) < actual {
		return nil, errors.New("abi: dynamic overflow")
	}
	return NewDecoder(d.xs[actual:]), nil
}
func (d *Decoder) ReadDynamicLength() (*Decoder, int, error) {
	offset, err := d.ReadBigUint()
	if err != nil {
		return nil, 0, err
	}
	actual := int(offset.Uint64())
	if len(d.xs) < actual {
		return nil, 0, errors.New("abi: dynamic overflow")
	}
	// hop over to the new one
	dec1 := NewDecoder(d.xs[actual:])
	l, err := dec1.ReadInt()
	if err != nil {
		return nil, 0, errors.New("abi: len unexpected EOF")
	}
	return NewDecoder(dec1.xs[32:]), l, nil
}

func (d *Decoder) Remaining() []byte {
	if d.cur > len(d.xs) {
		return nil
	}
	return d.xs[d.cur:]
}

func (d *Decoder) ReadString() (string, error) {
	offset, err := d.ReadBigUint()
	if err != nil {
		return "", err
	}
	actual := int(offset.Uint64())
	if len(d.xs) < actual {
		return "", errors.New("abi: dynamic overflow")
	}
	dec := NewDecoder(d.xs[actual:])
	l, err := dec.ReadUint()
	if err != nil {
		return "", err
	}
	bts, err := dec.ReadN(int(l))
	if err != nil {
		return "", err
	}
	return string(bts), nil
}
