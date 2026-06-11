package abi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallDecode(t *testing.T) {
	t.Run("typed params", func(t *testing.T) {
		sig, params := Call("add(uint256 1, uint256 2)").Decode()
		assert.EqualValues(t, "add(uint256,uint256)", sig)
		assert.EqualValues(t, []string{"1", "2"}, params)
	})
	t.Run("string param", func(t *testing.T) {
		sig, params := Call(`add(uint256 1, string "wowza")`).Decode()
		assert.EqualValues(t, "add(uint256,string)", sig)
		assert.EqualValues(t, []string{"1", `"wowza"`}, params)
	})
	t.Run("no params", func(t *testing.T) {
		sig, params := Call("foo()").Decode()
		assert.Equal(t, Signature("foo()"), sig)
		assert.Equal(t, []string{""}, params)
	})
	t.Run("too short", func(t *testing.T) {
		sig, params := Call("ab").Decode()
		assert.Equal(t, Signature(""), sig)
		assert.Nil(t, params)
	})
}

func TestSignatureTransfer(t *testing.T) {
	sig := Signature("transfer(address,uint256)")

	t.Run("Method", func(t *testing.T) {
		assert.Equal(t, "transfer", sig.Method())
	})
	t.Run("Args", func(t *testing.T) {
		assert.Equal(t, TypeName("(address,uint256)"), sig.Args())
	})
	t.Run("Hash known value", func(t *testing.T) {
		h := sig.Hash()
		assert.Equal(t, byte(0xa9), h[0])
		assert.Equal(t, byte(0x05), h[1])
		assert.Equal(t, byte(0x9c), h[2])
		assert.Equal(t, byte(0xbb), h[3])
	})
	t.Run("Hash caching", func(t *testing.T) {
		h1 := sig.Hash()
		h2 := sig.Hash()
		assert.Equal(t, h1, h2)
	})
	t.Run("Selector matches Hash prefix", func(t *testing.T) {
		h := sig.Hash()
		sel := sig.Selector()
		assert.Equal(t, [4]byte{h[0], h[1], h[2], h[3]}, sel)
	})
	t.Run("SelectorB matches Selector", func(t *testing.T) {
		sel := sig.Selector()
		assert.Equal(t, sel[:], sig.SelectorB())
	})
	t.Run("Fn matches Selector", func(t *testing.T) {
		sel := sig.Selector()
		assert.Equal(t, sel[:], sig.Fn())
	})
}

func TestSignatureMethod(t *testing.T) {
	tests := []struct {
		sig    Signature
		method string
	}{
		{"balanceOf(address)", "balanceOf"},
		{"approve(address,uint256)", "approve"},
		{"()", ""},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.method, tc.sig.Method(), "sig: %s", tc.sig)
	}
}

func TestSignatureArgs(t *testing.T) {
	tests := []struct {
		sig  Signature
		args TypeName
	}{
		{"balanceOf(address)", "(address)"},
		{"totalSupply()", "()"},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.args, tc.sig.Args(), "sig: %s", tc.sig)
	}
}

func TestSIGConstructor(t *testing.T) {
	t.Run("with args", func(t *testing.T) {
		assert.Equal(t, Signature("transfer(address,uint256)"), SIG("transfer", ADDRESS, UINT256))
	})
	t.Run("no args", func(t *testing.T) {
		assert.Equal(t, Signature("noArgs()"), SIG("noArgs"))
	})
}
