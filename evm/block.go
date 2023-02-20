package evm

import (
	"github.com/holiman/uint256"
)

func (s *InstructionSet) BlockHash() error {
	if len(s.s) < 1 {
		return ErrInvalidCode
	}
	s.s[0] = s.ctx.Block(s.s[0]).Hash().Clone()
	return nil
}

func (s *InstructionSet) Coinbase() error {
	return s.push(s.ctx.Block(nil).Coinbase().Word())
}

func (s *InstructionSet) Timestamp() error {
	return s.push(uint256.NewInt(s.ctx.Block(nil).Time()))
}

func (s *InstructionSet) Number() error {
	ans, _ := uint256.FromBig(s.ctx.Block(nil).Number())
	return s.push(ans)
}

func (s *InstructionSet) Difficulty() error {
	ans, _ := uint256.FromBig(s.ctx.Block(nil).Difficulty())
	return s.push(ans)
}

func (s *InstructionSet) GasLimit() error {
	return s.push(uint256.NewInt(s.ctx.Msg().Gas()))
}

func (s *InstructionSet) ChainID() error {
	return s.push(s.ctx.Block(nil).ChainID())
}

func (s *InstructionSet) SelfBalance() error {
	return s.push(s.ctx.Balance(s.ctx.Contract().Address()))
}
