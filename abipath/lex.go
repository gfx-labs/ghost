package abipath

import (
	"gfx.cafe/util/go/lexer"
)

func LexerState(l *lexer.Lex) lexer.StateFn {
	for {
		switch l.Cur() {
		case "":
		case ".", "/":
			l.Emit(lexer.SymbolToken)
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			l.TakeWhile("0123456789")
			l.Emit(lexer.IntegerToken)
		default:
			l.TakeUntil("./")
			l.Emit(lexer.StringToken)
		}
		if l.Next() == lexer.EOFRune {
			break
		}
	}
	l.Emit(lexer.EOFToken)
	return nil
}
