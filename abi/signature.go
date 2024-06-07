package abi

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var sigCache sync.Map

type Call string

func decodeCall(c string) (Signature, []string) {
	sigB := new(strings.Builder)
	if len(c) < 3 {
		return "", nil
	}
	methodIdx := strings.IndexByte(c, '(')
	sigB.WriteString(c[:methodIdx])
	sigB.WriteRune('(')
	out := make([]string, 0, 4)
	cutset := c[methodIdx+1 : len(c)-1]
	for idx, v := range strings.Split(cutset, ",") {
		if idx > 0 {
			sigB.WriteRune(',')
		}
		v = strings.TrimSpace(v)
		o := strings.SplitN(v, " ", 2)
		sigB.WriteString(strings.TrimSpace(o[0]))
		if len(o) > 1 {
			out = append(out, strings.TrimSpace(o[1]))
		} else {
			out = append(out, "")
		}
	}
	sigB.WriteRune(')')
	return Signature(sigB.String()), out
}
func (c Call) Decode() (Signature, []string) {
	return decodeCall(string(c))
}

// represents the string that is hashed for things like function signature and event signatures
type Signature string

func SIG(s string, t ...TypeName) Signature {
	return Signature(s + string(TUPLE(t...)))
}
func (s Signature) Method() string {
	ans := new(strings.Builder)
	for _, v := range s {
		if v == '(' {
			break
		}
		ans.WriteRune(v)
	}
	return ans.String()
}

func (s Signature) Args() TypeName {
	ans := new(strings.Builder)
	state := 0
	for _, v := range s {
		if v == '(' {
			state = 1
		}
		if state == 1 {
			ans.WriteRune(v)
		}
	}
	return TypeName(ans.String())
}

func (s Signature) Hash() common.Hash {
	if have, ok := sigCache.Load(string(s)); ok {
		return have.(common.Hash)
	}
	ans := common.BytesToHash(crypto.Keccak256([]byte(s)))
	sigCache.Store(string(s), ans)
	return ans
}
func (s Signature) SelectorB() []byte {
	ss := s.Selector()
	return ss[:]
}

func (s Signature) Selector() [4]byte {
	h := s.Hash()
	return [4]byte{h[0], h[1], h[2], h[3]}
}

func (s Signature) Fn() []byte {
	h := s.Hash()
	return []byte{h[0], h[1], h[2], h[3]}
}
