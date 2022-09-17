package evm

import (
	"github.com/holiman/uint256"
)

const MAX_SIZE = 1024

type Stack struct {
	s   []Word
	ctx CallContext
}

func (s *Stack) Allocate() {
	s.s = make([]Word, 0, s.MaxSize())
}
func (s *Stack) Size() int {
	return len(s.s)
}

func (s *Stack) MaxSize() int {
	return MAX_SIZE
}

func (s *Stack) Read(idx int) Word {
	if len(s.s) < idx {
		return *new(Word)
	}
	return s.s[idx]
}

func (s *Stack) Jump() error {
	s.ctx.Jump(s.s[0])
	return s.trim(1)
}
func (s *Stack) Jump1() error {
	s.ctx.Jump1(s.s[0], s.s[1])
	return s.trim(2)
}

func (s *Stack) Pc() error {
	return s.push(s.ctx.Counter())
}

func (s *Stack) MSize() error {
	return s.push(s.ctx.MemorySize())
}

func (s *Stack) JumpDest() error {
	return nil
}

func (s *Stack) PushN(l int, dat Word) error {
	bts := dat.Bytes()
	if len(bts) < l {
		l = len(bts)
	}
	rez := uint256.NewInt(0).SetBytes(bts[:l])
	return s.push(rez)
}
func (s *Stack) DupeN(idx int) error {
	// decrement idx because Dup starts counting at 1
	idx = idx - 1
	return s.push(s.s[idx])
}
func (s *Stack) SwapN(idx int) error {
	// decrement idx because Swap starts counting at 1
	idx = idx - 1
	s.s[0], s.s[idx] = s.s[idx], s.s[0]
	return nil
}

func (s *Stack) push(v Word) error {
	s.s = prepend(s.s, v)
	return nil
}

func (s *Stack) trim(i int) error {
	s.s = s.s[i:]
	return nil
}

func prepend[T any](x []T, y T) []T {
	x = append(x, *new(T))
	copy(x[1:], x)
	x[0] = y
	return x
}
