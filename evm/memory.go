package evm

import (
	"errors"
)

func (s *InstructionSet) Pop() error {
	return s.trim(1)
}

func (s *InstructionSet) MLoad() error {
	s.s[0] = s.ctx.MemoryAt(s.s[0])
	return nil
}

func (s *InstructionSet) MStore() error {
	if err := s.ctx.WriteMemory(s.s[0], s.s[1]); err != nil {
		return err
	}
	if err := s.trim(2); err != nil {
		return err
	}
	return nil
}

func (s *InstructionSet) MStore8() error {
	if err := s.ctx.WriteMemoryByte(s.s[0], s.s[1].Bytes()[0]); err != nil {
		return err
	}
	if err := s.trim(2); err != nil {
		return err
	}
	return nil
}
func (s *InstructionSet) SLoad() error {
	s.s[0] = s.ctx.Contract().StorageAt(s.s[0])
	return nil
}

func (s *InstructionSet) SStore() error {
	if err := s.ctx.Contract().WriteStorage(s.s[0], s.s[1]); err != nil {
		return err
	}
	if err := s.trim(2); err != nil {
		return err
	}
	return nil
}

func (s *InstructionSet) Revert() error {
	if err := s.trim(2); err != nil {
		return err
	}
	return errors.New("execution reverted")
}

func (s *InstructionSet) Invalid() error {
	return nil
}

func (s *InstructionSet) SelfDestruct() error {
	return s.trim(1)
}
