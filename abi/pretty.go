package abi

import (
	"encoding/hex"
	"strings"
)

// PrettyHex formats a byte slice as hex with one 32-byte word per line,
// separated by newlines. Useful for comparing ABI-encoded output in tests.
func PrettyHex(xs []byte) string {
	out := new(strings.Builder)
	cur := len(xs) % 32
	phex := hex.EncodeToString(xs[:cur])
	out.WriteString(phex)
	for {
		if cur+32 > len(xs) {
			out.WriteString(hex.EncodeToString(xs[cur:]))
			return out.String()
		} else {
			out.WriteRune('\n')
			out.WriteString(hex.EncodeToString(xs[cur:(cur + 32)]))
			cur = cur + 32
		}

	}
}
