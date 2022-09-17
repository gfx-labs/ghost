package evm

import (
	"github.com/holiman/uint256"
)

func (s *Stack) BlockHash() error {
	if len(s.s) < 1 {
		return ErrInvalidCode
	}
	s.s[0] = s.ctx.Block(s.s[0]).TxHash.Clone()
	return nil
}

func (s *Stack) Coinbase() error {
	return s.push(s.ctx.Block(nil).Coinbase.ToWord())
}

func (s *Stack) Timestamp() error {
	return s.push(uint256.NewInt(s.ctx.Block(nil).Time))
}

func (s *Stack) Number() error {
	ans, _ := uint256.FromBig(s.ctx.Block(nil).Number)
	return s.push(ans)
}

func (s *Stack) Difficulty() error {
	ans, _ := uint256.FromBig(s.ctx.Block(nil).Difficulty)
	return s.push(ans)
}

func (s *Stack) GasLimit() error {
	return s.push(uint256.NewInt(s.ctx.Block(nil).GasLimit))
}

func (s *Stack) ChainID() error {
	return s.push(s.ctx.Block(nil).ChainID)
}

func (s *Stack) SelfBalance() error {
	return s.push(s.ctx.Balance(s.ctx.Contract().Address()))
}
