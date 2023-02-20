package evm

import (
	"github.com/holiman/uint256"
)

const MAX_SIZE = 1024

type InstructionSet struct {
	s   []Word
	ctx Engine
}

func NewExecutor(engine Engine) *InstructionSet {
	return &InstructionSet{
		ctx: engine,
	}
}

// copies stack to new with executor
func (s *InstructionSet) WithEngine(engine Engine) *InstructionSet {
	return &InstructionSet{s: s.Copy(), ctx: engine}
}

func (s *InstructionSet) Copy() []Word {
	sCopy := make([]*uint256.Int, len(s.s))
	for i, i2 := range s.s {
		sCopy[i] = i2
	}
	return sCopy
}

func (s *InstructionSet) View(fn func([]Word) error) error {
	return fn(s.s)
}

func (s *InstructionSet) Allocate() {
	s.s = make([]Word, 0, s.MaxSize())
}
func (s *InstructionSet) Size() int {
	return len(s.s)
}

func (s *InstructionSet) MaxSize() int {
	return MAX_SIZE
}

func (s *InstructionSet) Read(idx int) Word {
	if len(s.s) < idx {
		return *new(Word)
	}
	return s.s[idx]
}

func (s *InstructionSet) Jump() error {
	s.ctx.Jump(s.s[0])
	return s.trim(1)
}
func (s *InstructionSet) Jump1() error {
	s.ctx.Jump1(s.s[0], s.s[1])
	return s.trim(2)
}

func (s *InstructionSet) Pc() error {
	return s.push(s.ctx.Counter())
}

func (s *InstructionSet) MSize() error {
	return s.push(s.ctx.MemorySize())
}

func (s *InstructionSet) JumpDest() error {
	return nil
}

func (s *InstructionSet) PushN(l int, dat Word) error {
	bts := dat.Bytes()
	if len(bts) < l {
		l = len(bts)
	}
	rez := uint256.NewInt(0).SetBytes(bts[:l])
	return s.push(rez)
}
func (s *InstructionSet) DupeN(idx int) error {
	// decrement idx because Dup starts counting at 1
	idx = idx - 1
	return s.push(s.s[idx])
}
func (s *InstructionSet) SwapN(idx int) error {
	// decrement idx because Swap starts counting at 1
	idx = idx - 1
	s.s[0], s.s[idx] = s.s[idx], s.s[0]
	return nil
}

func (s *InstructionSet) push(v Word) error {
	s.s = prepend(s.s, v)
	return nil
}

func (s *InstructionSet) trim(i int) error {
	s.s = s.s[i:]
	return nil
}

func prepend[T any](x []T, y T) []T {
	x = append(x, *new(T))
	copy(x[1:], x)
	x[0] = y
	return x
}
