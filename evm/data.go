package evm

import (
	"github.com/ethereum/go-ethereum/common"
)

func (s *Stack) Address() error {
	return s.push(AddressToWord(s.ctx.Contract().Address()))
}

func (s *Stack) Balance() error {
	s.s[0] = s.ctx.Balance(common.BytesToAddress(s.s[0].Bytes()))
	return nil
}

func (s *Stack) Origin() error {
	return s.push(AddressToWord(s.ctx.Txn().From))
}

func (s *Stack) Caller() error {
	return s.push(AddressToWord(s.ctx.Caller()))
}

func (s *Stack) CallValue() error {
	return s.push(s.ctx.Txn().Value)
}

func (s *Stack) CallDataLoad() error {
	return s.push(s.ctx.CallData(s.s[0].Uint64()))
}

func (s *Stack) CallDataSize() error {
	return s.push(s.ctx.CallDataSize())
}

func (s *Stack) CallDataCopy() error {
	if err := s.ctx.CallDataCopy(s.s[0], s.s[1], s.s[2]); err != nil {
		return err
	}
	return s.trim(3)
}

func (s *Stack) CodeSize() error {
	return s.push(s.ctx.CodeSize())
}

func (s *Stack) CodeCopy() error {
	if err := s.ctx.CodeCopy(s.s[0], s.s[1], s.s[2]); err != nil {
		return err
	}
	return s.trim(3)
}

func (s *Stack) GasPrice() error {
	return s.push(s.ctx.Txn().GasPrice)
}

func (s *Stack) ExtCodeSize() error {
	s.s[0] = s.ctx.ExtCodeSize(common.BytesToAddress(s.s[0].Bytes()))
	return nil
}

func (s *Stack) ExtCodeCopy() error {
	if err := s.ctx.ExtCodeCopy(s.s[0], s.s[1], s.s[2], s.s[3]); err != nil {
		return err
	}
	return s.trim(4)
}

func (s *Stack) ReturnDataSize() error {
	return s.push(s.ctx.ReturnDataSize())
}

func (s *Stack) ReturnDataCopy() error {
	if err := s.ctx.ReturnDataCopy(s.s[0], s.s[1], s.s[2]); err != nil {
		return err
	}
	return s.trim(3)
}

func (s *Stack) ExtCodeHash() error {
	return s.push(s.ctx.ExtCodeHash(common.BytesToAddress(s.s[0].Bytes())))
}
