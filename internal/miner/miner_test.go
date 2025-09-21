package miner_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/miner"
)

func TestMiner(t *testing.T) {
	miner1 := blockchain.MustGenerateTestAddress(t)

	l, err := blockchain.NewLedger(10)
	if err != nil {
		t.Fatal(err)
	}

	m := miner.NewMiner(l, miner1.PublicKey())

	m.Mine(8)

	time.Sleep(time.Second)

	fmt.Print(l.Length())
}
