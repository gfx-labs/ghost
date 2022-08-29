package copypasta

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
)

func TestFilterLogs(t *testing.T) {
	client, err := Dial("https://mainnet.rpc.gfx.xyz")
	if err != nil {
		t.Fatal(err)
	}
	logs, err := client.ErigonFilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: big.NewInt(1400000),
		ToBlock:   big.NewInt(1401000),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range logs {
		if v.Timestamp == 0 {
			t.Errorf("log timestamp 0: %+v", v)
		}
	}

}
