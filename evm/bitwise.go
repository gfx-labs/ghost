package evm

func (s *InstructionSet) And() error {
	s.s[1].And(s.s[1], s.s[0])
	return s.trim(1)

}

func (s *InstructionSet) Or() error {
	s.s[1].Or(s.s[1], s.s[0])
	return s.trim(1)

}

func (s *InstructionSet) Xor() error {
	s.s[1].Xor(s.s[1], s.s[0])
	return s.trim(1)
}

func (s *InstructionSet) Not() error {
	s.s[0].Not(s.s[0])
	return nil
}

func (s *InstructionSet) Byte() error {
	s.s[1].Byte(s.s[0])
	return s.trim(1)
}

func (s *InstructionSet) Shl() error {
	s.s[1].Lsh(s.s[1], uint(s.s[0].Uint64()))
	return s.trim(1)
}

func (s *InstructionSet) Shr() error {
	s.s[1].Rsh(s.s[1], uint(s.s[0].Uint64()))
	return s.trim(1)
}

func (s *InstructionSet) Sar() error {
	s.s[1].SRsh(s.s[1], uint(s.s[0].Uint64()))
	return s.trim(1)
}
