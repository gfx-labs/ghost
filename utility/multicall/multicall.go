package multicall

import (
	"gfx.cafe/open/ghost/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type Call struct {
	Target common.Address
	Data   []byte
}

type Call3 struct {
	Target       common.Address
	AllowFailure bool
	Data         []byte
}

type Call3Value struct {
	Target       common.Address
	AllowFailure bool
	Value        uint256.Int
	Data         []byte
}

var tryAggregateSignature = abi.SIG("tryAggregate", abi.BOOL, abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.BYTES))).Fn()

func EncodeTryAggregate(requireSuccess bool, calls []Call) []byte {
	b := new(abi.Builder)
	b = b.Bool(requireSuccess)
	b = b.EnterDynamicArray()
	for _, v := range calls {
		b.EnterTuple().Address(v.Target).Bytes(v.Data).Exit()
	}
	b = b.Exit()
	data := b.Finish(tryAggregateSignature)
	return data
}

var aggregate3Signature = abi.SIG("aggregate3", abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.BOOL, abi.BYTES))).Fn()

func EncodeAggregate3(calls []Call3) []byte {
	b := new(abi.Builder)
	b = b.EnterDynamicArray()
	for _, v := range calls {
		b.EnterTuple().Address(v.Target).Bool(v.AllowFailure).Bytes(v.Data).Exit()
	}
	b = b.Exit()
	data := b.Finish(aggregate3Signature)
	return data
}

var aggregate3ValueSignature = abi.SIG("aggregate3Value", abi.SLICE(abi.TUPLE(abi.ADDRESS, abi.BOOL, abi.UINT256, abi.BYTES))).Fn()

func EncodeAggregate3Value(calls []Call3Value) []byte {
	b := new(abi.Builder)
	b = b.EnterDynamicArray()
	for _, v := range calls {
		b.EnterTuple().Address(v.Target).Bool(v.AllowFailure).Uint256(&v.Value).Bytes(v.Data).Exit()
	}
	b = b.Exit()
	data := b.Finish(aggregate3ValueSignature)
	return data
}

type CallResult struct {
	Success bool
	Data    []byte
	Err     error
}

func DecodeResult(data []byte) ([]*CallResult, error) {
	dec := abi.NewDecoder(data)
	d, l, err := dec.DynamicLength()
	if err != nil {
		return nil, err
	}
	abiRes := []*CallResult{}
	for i := 0; i < l; i++ {
		bts := &CallResult{}
		nd, err := d.Dynamic()
		if err != nil {
			return nil, err
		}
		bts.Success, err = nd.Bool()
		if err != nil {
			return nil, err
		}
		bts.Data, err = nd.Bytes()
		if err != nil {
			return nil, err
		}
		abiRes = append(abiRes, bts)
	}
	return abiRes, nil
}
