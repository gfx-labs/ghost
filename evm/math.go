package evm

import (
	"github.com/holiman/uint256"
)

var int256min, _ = uint256.FromHex("0x8000000000000000000000000000000000000000000000000000000000000000")
var neg1, _ = uint256.FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

func (s *InstructionSet) Add() error {
	s.s[1].Add(s.s[1], s.s[0])
	return s.trim(1)
}
func (s *InstructionSet) Mul() error {
	s.s[1].Mul(s.s[1], s.s[0])
	return s.trim(1)
}

func (s *InstructionSet) Sub() error {
	s.s[1].Sub(s.s[0], s.s[1])
	return s.trim(1)
}

func (s *InstructionSet) Div() error {
	if s.s[1].IsZero() {
		s.s[1].Clear()
	} else {
		s.s[1].Div(s.s[0], s.s[1])
	}
	return s.trim(1)
}

func (s *InstructionSet) SDiv() error {
	if s.s[1].IsZero() {
		s.s[1].Clear()
	} else {
		if s.s[0].Eq(int256min) && s.s[1].Eq(neg1) {
			//	storage.OverwriteWord(s.s[1], int256min)
		} else {
			s.s[1].SDiv(s.s[0], s.s[1])
		}
	}
	return s.trim(1)
}
func (s *InstructionSet) Mod() error {
	if s.s[1].IsZero() {
		s.s[1].Clear()
	} else {
		s.s[1].Mod(s.s[0], s.s[1])
	}
	return s.trim(1)
}

func (s *InstructionSet) SMod() error {
	if s.s[1].IsZero() {
		s.s[1].Clear()
	} else {
		s.s[1].SMod(s.s[0], s.s[1])
	}
	return s.trim(1)
}

func (s *InstructionSet) AddMod() error {
	if err := s.Add(); err != nil {
		return err
	}
	return s.Mod()
}

func (s *InstructionSet) MulMod() error {
	if err := s.Mul(); err != nil {
		return err
	}
	return s.Mod()
}

func (s *InstructionSet) Exp() error {
	s.s[1].Exp(s.s[0], s.s[1])
	return s.trim(1)
}

func (s *InstructionSet) SignExtend() error {
	s.s[1].ExtendSign(s.s[1], s.s[0])
	return s.trim(1)
}

func (s *InstructionSet) Lt() error {
	s.s[1].Clear()
	if s.s[0].Lt(s.s[1]) {
		s.s[1][0] = 1
	}
	return s.trim(1)
}

func (s *InstructionSet) Gt() error {
	s.s[1].Clear()
	if s.s[0].Gt(s.s[1]) {
		s.s[1][0] = 1
	}
	return s.trim(1)
}

func (s *InstructionSet) Slt() error {
	s.s[1].Clear()
	if s.s[0].Slt(s.s[1]) {
		s.s[1][0] = 1
	}
	return s.trim(1)
}

func (s *InstructionSet) Sgt() error {
	s.s[1].Clear()
	if s.s[0].Sgt(s.s[1]) {
		s.s[1][0] = 1
	}
	return s.trim(1)
}

func (s *InstructionSet) Eq() error {
	s.s[1].Clear()
	if s.s[0].Eq(s.s[1]) {
		s.s[1][0] = 1
	}
	return s.trim(1)
}

func (s *InstructionSet) IsZero() error {
	s.s[1].Clear()
	if s.s[0].IsZero() {
		s.s[0][0] = 1
	}
	return nil
}
