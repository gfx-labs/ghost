// Package testutil provides helpers for testing ABI encoding and decoding.
package testutil

import (
	"encoding/hex"
	"strings"

	"github.com/gfx-labs/ghost/abi"
)

var hexReplacer = strings.NewReplacer(
	" ", "",
	"\t", "",
	"\n", "",
	"\r", "",
	"\v", "",
	"\f", "",
)

// HexDecoder creates an abi.Decoder from a hex string, stripping whitespace
// and optional 0x/0X prefix. Panics on invalid hex.
func HexDecoder(s string) *abi.Decoder {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = hexReplacer.Replace(s)
	ans, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return abi.NewDecoder(ans)
}

// HexBytes decodes a hex string (with optional 0x prefix and whitespace)
// into raw bytes. Panics on invalid hex.
func HexBytes(s string) []byte {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = hexReplacer.Replace(s)
	ans, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return ans
}
