package copypasta

import (
	"context"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
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
func TestReceipt(t *testing.T) {
	client, err := Dial("https://mainnet.rpc.gfx.xyz")
	if err != nil {
		t.Fatal(err)
	}
	receipts, err := client.ErigonGetReceiptsByHash(context.TODO(), common.HexToHash("0xe9f3dc5f34ca7a507abc1e290d2165ca297ac1b34e4b5bd89615177fc3174726"))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range receipts {
		for _, vv := range v.Logs {
			log.Println(vv)
		}
	}
}
