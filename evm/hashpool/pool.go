package hashpool

import (
	"hash"
	"sync"

	"golang.org/x/crypto/sha3"
)

type KeccakHashPool struct {
	p sync.Pool
}
type KeccackHasher interface {
	hash.Hash
	Read([]byte) (int, error)
}

type keccackState struct {
	p *KeccakHashPool
	KeccackHasher
}

func (h *keccackState) Return() {
	h.p.Put(h)
}

func NewKeccack() *KeccakHashPool {
	out := &KeccakHashPool{}
	out.p = sync.Pool{}
	out.p.New = func() any {
		return &keccackState{
			p:             out,
			KeccackHasher: sha3.NewLegacyKeccak256().(KeccackHasher),
		}
	}
	return out
}

func (p *KeccakHashPool) Get() *keccackState {
	ans := p.p.Get().(*keccackState)
	ans.Reset()
	return ans
}

func (p *KeccakHashPool) Put(k *keccackState) {
	p.Put(k)
}
