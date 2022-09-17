package evm

import (
	"log"
	"testing"

	"github.com/holiman/uint256"
)

func TestPushBytes(t *testing.T) {
	stack := &Stack{}
	stack.PushN(32, uint256.NewInt(8))
	stack.PushN(32, uint256.NewInt(8))
	stack.PushN(32, uint256.NewInt(8))
	stack.PushN(32, uint256.NewInt(8))
	stack.Add()
	stack.Add()
	stack.Add()
	if rez := stack.Read(0)[0]; rez != 32 {
		log.Fatalf("expected %d got %d", 32, rez)
	}
	stack.PushN(32, uint256.NewInt(100))
	stack.Mul()
	if rez := stack.Read(0).Uint64(); rez != 3200 {
		log.Fatalf("expected %d got %d", 3200, rez)
	}
}
