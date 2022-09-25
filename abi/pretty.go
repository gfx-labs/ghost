package abi

import (
	"encoding/hex"
	"strings"
)

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
