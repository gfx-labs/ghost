package abipath

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"gfx.cafe/util/go/lexer"
)

func word(v byte) []byte {
	var w [32]byte
	w[31] = v
	return w[:]
}

func TestLexerState(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []lexer.Token
	}{
		{"empty", "", []lexer.Token{{Type: lexer.EOFToken}}},
		{"dot", ".", []lexer.Token{{Type: lexer.SymbolToken, Value: "."}, {Type: lexer.EOFToken}}},
		{"slash", "/", []lexer.Token{{Type: lexer.SymbolToken, Value: "/"}, {Type: lexer.EOFToken}}},
		{"digit", "5", []lexer.Token{{Type: lexer.IntegerToken, Value: "5"}, {Type: lexer.EOFToken}}},
		{"multi digit", "123", []lexer.Token{{Type: lexer.IntegerToken, Value: "123"}, {Type: lexer.EOFToken}}},
		{"string", "foo", []lexer.Token{{Type: lexer.StringToken, Value: "foo"}, {Type: lexer.EOFToken}}},
		{"path dot", "a.b", []lexer.Token{{Type: lexer.StringToken, Value: "a"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.StringToken, Value: "b"}, {Type: lexer.EOFToken}}},
		{"path slash", "a/b", []lexer.Token{{Type: lexer.StringToken, Value: "a"}, {Type: lexer.SymbolToken, Value: "/"}, {Type: lexer.StringToken, Value: "b"}, {Type: lexer.EOFToken}}},
		{"complex", ".0.1/2", []lexer.Token{{Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "0"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "1"}, {Type: lexer.SymbolToken, Value: "/"}, {Type: lexer.IntegerToken, Value: "2"}, {Type: lexer.EOFToken}}},
		{"all digits", "0.1.2.3.4.5.6.7.8.9", []lexer.Token{{Type: lexer.IntegerToken, Value: "0"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "1"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "2"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "3"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "4"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "5"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "6"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "7"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "8"}, {Type: lexer.SymbolToken, Value: "."}, {Type: lexer.IntegerToken, Value: "9"}, {Type: lexer.EOFToken}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input, LexerState)
			var got []lexer.Token
			l.ConsumeWithUntilErr(func(tok *lexer.Token) error {
				got = append(got, *tok)
				return nil
			})

			require.Len(t, got, len(tc.expect), "token count mismatch")
			for i, want := range tc.expect {
				require.Equal(t, want.Type, got[i].Type, "token[%d] type", i)
				if want.Value != "" {
					require.Equal(t, want.Value, got[i].Value, "token[%d] value", i)
				}
			}
		})
	}
}

func TestPoint(t *testing.T) {
	// data layout: word(32)=offset to dynamic, word(64)=length, then payload
	data := bytes.Join([][]byte{
		word(64),       // offset 0: points to byte 64
		word(99),       // offset 32: filler
		word(3),        // offset 64: length = 3
		{1, 2, 3, 4},   // offset 96: payload
	}, nil)

	tests := []struct {
		name   string
		path   string
		input  []byte
		expect []byte
		err    bool
	}{
		{"empty path", "", data, data, false},
		{"dot dynamic", ".", data, data[64:], false},
		{"slash dynamic length", "/", data, data[96:], false},
		{"skip one word", "1", data, data[32:], false},
		{"skip two words", "2", data, data[64:], false},
		{"dot then skip", ".1", data, data[96:], false},
		{"slash then skip", "/0", data, data[96:], false},
		{"dot error", ".", word(255), nil, true},
		{"slash error", "/", word(32), nil, true},
		{"read error", "1", nil, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Point(tc.path, tc.input)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}
