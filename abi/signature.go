package abi

import (
	"strings"

	"gfx.cafe/util/go/generic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var sigCache generic.Map[string, common.Hash]

type Call string

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
		return have
	}
	ans := common.BytesToHash(crypto.Keccak256([]byte(s)))
	sigCache.Store(string(s), ans)
	return ans
}

func (s Signature) Selector() [4]byte {
	h := s.Hash()
	return [4]byte{h[0], h[1], h[2], h[3]}
}

func (s Signature) Fn() []byte {
	h := s.Hash()
	return []byte{h[0], h[1], h[2], h[3]}
}
