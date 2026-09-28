package lexer

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func charState(l *Lex) StateFn {
	for l.Next() != EOFRune {
		l.Emit(StringToken)
	}
	l.Emit(EOFToken)
	return nil
}

func TestTokenInt(t *testing.T) {
	n, err := Token{Type: IntegerToken, Value: "42"}.Int()
	require.NoError(t, err)
	require.Equal(t, 42, n)

	_, err = Token{Type: IntegerToken, Value: "99999999999999999999"}.Int()
	require.Error(t, err)

	_, err = Token{Type: StringToken, Value: "1"}.Int()
	require.Error(t, err)
}

func TestConsumeWithUntilErrNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	stop := errors.New("stop")
	for i := 0; i < 50; i++ {
		l := New("abcdefghijklmnopqrstuvwxyz", charState)
		err := l.ConsumeWithUntilErr(func(*Token) error { return stop })
		require.ErrorIs(t, err, stop)
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	require.LessOrEqual(t, runtime.NumGoroutine(), before)
}

func TestErrorNoHandler(t *testing.T) {
	l := &Lex{}
	require.NotPanics(t, func() { l.Error("bad") })
	require.EqualError(t, l.Err, "bad")
}
