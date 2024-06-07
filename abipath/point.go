package abipath

import (
	"gfx.cafe/open/ghost/abi"
	"gfx.cafe/util/go/lexer"
)

func Point(p string, xs []byte) ([]byte, error) {
	dec := abi.NewDecoder(xs)
	l := lexer.New(p, LexerState)
	err := l.ConsumeWithUntilErr(func(tok *lexer.Token) (err error) {
		switch tok.Type {
		case lexer.EOFToken:
			return nil
		case lexer.EmptyToken:
		case lexer.SymbolToken:
			switch tok.Value {
			case ".":
				dec, err = dec.Dynamic()
				if err != nil {
					return err
				}
			case "/":
				dec, _, err = dec.DynamicLength()
				if err != nil {
					return err
				}
			}
		case lexer.IntegerToken:
			_, err := dec.ReadN(tok.Int() * 32)
			if err != nil {
				return err
			}
		case lexer.StringToken:
			// nothing for now...
			// in the future we can try to parse an abi or something?
			return nil
		default:
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dec.Remaining(), nil
}
