package lexer

type runeStack struct {
	xs []rune
}

func (s *runeStack) push(r rune) {
	s.xs = append(s.xs, r)
}

func (s *runeStack) pop() (o rune) {
	if len(s.xs) == 0 {
		return EOFRune
	}
	s.xs, o = s.xs[:len(s.xs)-1], s.xs[len(s.xs)-1]
	return
}

func (s *runeStack) clear() {
	s.xs = s.xs[:0]
}
