package evm

import (
	"github.com/holiman/uint256"
)

func (s *InstructionSet) LogN(idx int) error {
	topics := make([]*uint256.Int, idx)
	for i := 2; i < (idx + 2); i++ {
		topics[i-2] = s.s[i]
	}
	if err := s.ctx.WriteLog(s.s[0], s.s[1], topics); err != nil {
		return err
	}
	return s.trim(idx + 2)
}
