package abi

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var sigCache sync.Map

// Call represents a human-readable function call string with embedded argument
// values, such as "add(uint256 1, uint256 2)". Use [Call.Decode] to parse it
// into a [Signature] and parameter values.
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
// Decode parses the call string into a canonical signature and the
// string representations of each argument value.
//
//	sig, params := Call("add(uint256 1, uint256 2)").Decode()
//	// sig = "add(uint256,uint256)", params = ["1", "2"]
func (c Call) Decode() (Signature, []string) {
	return decodeCall(string(c))
}

// Signature is a canonical Solidity function or event signature string,
// such as "transfer(address,uint256)". Its [Signature.Hash] method computes
// the Keccak-256 hash, and [Signature.Selector] / [Signature.Fn] extract
// the 4-byte function selector.
type Signature string

// SIG constructs a Signature from a method name and argument types:
//
//	SIG("transfer", ADDRESS, UINT256) => "transfer(address,uint256)"
//	SIG("totalSupply")                => "totalSupply()"
func SIG(s string, t ...TypeName) Signature {
	return Signature(s + string(TUPLE(t...)))
}
// Method returns the function name portion before the opening parenthesis.
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

// Args returns the argument portion as a tuple TypeName, e.g. "(address,uint256)".
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

// Hash returns the Keccak-256 hash of the signature string.
// Results are cached in a global sync.Map for repeated lookups.
func (s Signature) Hash() common.Hash {
	if have, ok := sigCache.Load(string(s)); ok {
		return have.(common.Hash)
	}
	ans := common.BytesToHash(crypto.Keccak256([]byte(s)))
	sigCache.Store(string(s), ans)
	return ans
}
// SelectorB returns the 4-byte function selector as a byte slice.
func (s Signature) SelectorB() []byte {
	ss := s.Selector()
	return ss[:]
}

// Selector returns the 4-byte function selector as a fixed-size array.
func (s Signature) Selector() [4]byte {
	h := s.Hash()
	return [4]byte{h[0], h[1], h[2], h[3]}
}

// Fn returns the 4-byte function selector as a byte slice, suitable for
// passing to [Builder.Finish] as a calldata prefix.
func (s Signature) Fn() []byte {
	h := s.Hash()
	return []byte{h[0], h[1], h[2], h[3]}
}
