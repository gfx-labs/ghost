package lexer

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// a lexer based off
// https://go.dev/talks/2011/lex.slide

type TokenType int

const (
	EOFRune rune = -1

	EOFToken TokenType = (math.MaxInt - iota)
	EmptyToken
	StringToken
	IntegerToken
	SymbolToken
)

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) Int() (i int) {
	if t.Type == IntegerToken {
		i, _ = strconv.Atoi(t.Value)
		return
	}
	return
}

type StateFn func(*Lex) StateFn

type Lex struct {
	src        string
	start, pos int

	tokens chan Token
	rewind runeStack

	Err        error
	ErrHandler func(e string)
}

type LexOption = func(*Lex)

// New creates a lexer
func New(src string, start StateFn, opts ...LexOption) *Lex {
	l := &Lex{
		src: src,
	}
	l.tokens = make(chan Token, len(l.src)/2+1)
	for _, v := range opts {
		v(l)
	}
	go l.run(start)
	return l
}

// calls the function for every token until done
// this is just a helper for NextToken
func (l *Lex) ConsumeWith(fn func(tok *Token)) {
	for {
		tok, done := l.NextToken()
		if done {
			break
		}
		fn(tok)
	}
}

// calls the function for every token until done or error
// this is just a helper for NextToken
func (l *Lex) ConsumeWithUntilErr(fn func(tok *Token) error) error {
	for {
		tok, done := l.NextToken()
		if done {
			break
		}
		err := fn(tok)
		if err != nil {
			return err
		}
	}
	return nil
}

// returns a token, and whether or not is done
func (l *Lex) NextToken() (tok *Token, done bool) {
	if tok, done := <-l.tokens; done {
		return &tok, false
	}
	return nil, true
}

// returns current buffer
func (l *Lex) Cur() string {
	return l.src[l.start:l.pos]
}

// emits the current buffer as token type into the channel
func (l *Lex) Emit(t TokenType) {
	tok := Token{
		Type:  t,
		Value: l.Cur(),
	}
	l.tokens <- tok
	l.Ignore()
}

// clears the rewind stack and then sets the current beginning pos
// to the current pos in the src, this is what happens after emit
func (l *Lex) Ignore() {
	l.rewind.clear()
	l.start = l.pos
}

// peek the next rune
func (l *Lex) Peek() rune {
	return l.next(false)
}

// next pulls the next rune from the Lexer and returns it, moving the pos forward in the src.
func (l *Lex) Next() rune {
	return l.next(true)
}

// reads tokens until a token not in the given string is encountered.
func (l *Lex) TakeWhile(chars string) {
	l.take(chars, true)
}

// reads tokens until a token in the given string is encountered.
func (l *Lex) TakeUntil(chars string) {
	l.take(chars, false)
}

// take the last rune read (if any) and rewind back.
// rewinds can occur more than once per call to Next but you can never rewind past the
// last point a token was emitted.
func (l *Lex) Rewind() {
	r := l.rewind.pop()
	if r > EOFRune {
		l.pos = l.pos - utf8.RuneLen(r)
		if l.pos < l.start {
			l.pos = l.start
		}
	}
}

// either emits an error and calls the error handler, or panics if there is no error handler
func (l *Lex) Error(e string) {
	if l.ErrHandler == nil {
		panic(e)
	}
	l.Err = errors.New(e)
	l.ErrHandler(e)
}

// recursive reader
func (l *Lex) run(startState StateFn) {
	for state := startState; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

// reads tokens until a token not in the given string is encountered.
func (l *Lex) take(chars string, while bool) {
	r := l.next(true)
	for strings.ContainsRune(chars, r) == while && r != EOFRune {
		r = l.next(true)
	}
	if r == EOFRune {
		return
	}
	l.Rewind() // last next wasn't a match
}

// actual implementation of peek/next. inc == true if its not a peek
func (l *Lex) next(inc bool) rune {
	str := l.src[l.pos:]
	if len(str) == 0 {
		return EOFRune
	}
	r, s := utf8.DecodeRuneInString(str)
	if inc {
		l.pos += s
		l.rewind.push(r)
	}
	return r
}
