package evm

func (s *InstructionSet) Address() error {
	return s.push(s.ctx.Contract().Address().Word())
}

func (s *InstructionSet) Balance() error {
	s.s[0] = s.ctx.Balance(s.s[0].Bytes20())
	return nil
}

func (s *InstructionSet) Origin() error {
	return s.push(s.ctx.Msg().From().Word())
}

func (s *InstructionSet) Caller() error {
	return s.push(s.ctx.Caller().Word())
}

func (s *InstructionSet) CallValue() error {
	return s.push(s.ctx.Msg().Value())
}

func (s *InstructionSet) CallDataLoad() error {
	return s.push(s.ctx.CallData(s.s[0].Uint64()))
}

func (s *InstructionSet) CallDataSize() error {
	return s.push(s.ctx.CallDataSize())
}

func (s *InstructionSet) CallDataCopy() error {
	if err := s.ctx.CallDataCopy(s.s[0], s.s[1], s.s[2]); err != nil {
		return err
	}
	return s.trim(3)
}

func (s *InstructionSet) CodeSize() error {
	return s.push(s.ctx.CodeSize())
}

func (s *InstructionSet) CodeCopy() error {
	if err := s.ctx.CodeCopy(s.s[0], s.s[1], s.s[2]); err != nil {
		return err
	}
	return s.trim(3)
}

func (s *InstructionSet) GasPrice() error {
	return s.push(s.ctx.Msg().GasPrice())
}

func (s *InstructionSet) ExtCodeSize() error {
	s.s[0] = s.ctx.ExtCodeSize(s.s[0].Bytes20())
	return nil
}

func (s *InstructionSet) ExtCodeCopy() error {
	if err := s.ctx.ExtCodeCopy(s.s[0], s.s[1], s.s[2], s.s[3]); err != nil {
		return err
	}
	return s.trim(4)
}

func (s *InstructionSet) ReturnDataSize() error {
	return s.push(s.ctx.ReturnDataSize())
}

func (s *InstructionSet) ReturnDataCopy() error {
	if err := s.ctx.ReturnDataCopy(s.s[0], s.s[1], s.s[2]); err != nil {
		return err
	}
	return s.trim(3)
}

func (s *InstructionSet) ExtCodeHash() error {
	return s.push(s.ctx.ExtCodeHash(s.s[0].Bytes20()))
}
