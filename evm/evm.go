package evm

type EVMv0 struct {
}

func Execute(s *InstructionSet) error {
	contractCreation := s.ctx.Msg().To() == nil
	if contractCreation {
		return s.ctx.CreateContract()
	}
	return nil
}
