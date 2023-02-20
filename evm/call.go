package evm

func (s *InstructionSet) Create() error {
	addr, err := s.ctx.Create(s.s[0], s.s[1], s.s[2])
	if err != nil {
		return err
	}
	if err := s.trim(3); err != nil {
		return err
	}
	return s.push(addr.Word())
}
func (s *InstructionSet) Create2() error {
	addr, err := s.ctx.Create2(s.s[0], s.s[1], s.s[2], s.s[3])
	if err != nil {
		return err
	}
	if err := s.trim(4); err != nil {
		return err
	}
	return s.push(addr.Word())
}
func (s *InstructionSet) DelegateCall() error {
	rez, err := s.ctx.DelegateCall(s.s[0], s.s[1], s.s[2], s.s[3], s.s[4], s.s[5])
	if err != nil {
		return err
	}
	if err := s.trim(6); err != nil {
		return err
	}
	return s.push(rez)
}
func (s *InstructionSet) StaticCall() error {
	rez, err := s.ctx.DelegateCall(s.s[0], s.s[1], s.s[2], s.s[3], s.s[4], s.s[5])
	if err != nil {
		return err
	}
	if err := s.trim(6); err != nil {
		return err
	}
	return s.push(rez)
}
func (s *InstructionSet) Call() error {
	rez, err := s.ctx.Call(s.s[0], s.s[1], s.s[2], s.s[3], s.s[4], s.s[5], s.s[6])
	if err != nil {
		return err
	}
	if err := s.trim(7); err != nil {
		return err
	}
	return s.push(rez)
}

func (s *InstructionSet) CallCode() error {
	rez, err := s.ctx.CallCode(s.s[0], s.s[1], s.s[2], s.s[3], s.s[4], s.s[5], s.s[6])
	if err != nil {
		return err
	}
	if err := s.trim(7); err != nil {
		return err
	}
	return s.push(rez)
}

func (s *InstructionSet) Return() error {
	if err := s.Return(); err != nil {
		return err
	}
	return nil
}
