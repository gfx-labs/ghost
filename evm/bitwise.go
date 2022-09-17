package evm

func (s *Stack) And() error {
	s.s[1].And(s.s[1], s.s[0])
	return s.trim(1)

}

func (s *Stack) Or() error {
	s.s[1].Or(s.s[1], s.s[0])
	return s.trim(1)

}

func (s *Stack) Xor() error {
	s.s[1].Xor(s.s[1], s.s[0])
	return s.trim(1)
}

func (s *Stack) Not() error {
	s.s[0].Not(s.s[0])
	return nil
}

func (s *Stack) Byte() error {
	s.s[1].Byte(s.s[0])
	return s.trim(1)
}

func (s *Stack) Shl() error {
	s.s[1].Lsh(s.s[1], uint(s.s[0].Uint64()))
	return s.trim(1)
}

func (s *Stack) Shr() error {
	s.s[1].Rsh(s.s[1], uint(s.s[0].Uint64()))
	return s.trim(1)
}

func (s *Stack) Sar() error {
	s.s[1].SRsh(s.s[1], uint(s.s[0].Uint64()))
	return s.trim(1)
}
