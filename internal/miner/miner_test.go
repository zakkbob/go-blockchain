package miner_test

import (
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/miner"
)

func TestMiner(t *testing.T) {
	miner1 := blockchain.MustGenerateTestAddress(t)

	l, err := blockchain.NewLedger(2)
	if err != nil {
		t.Fatal(err)
	}

	b := l.ConstructNextBlock(map[[32]byte]blockchain.Transaction{}, miner1.PublicKey())
	m := miner.NewMiner(miner1.PublicKey(), 8)
	m.SetTargetBlock(b)
	mined := <-m.MinedBlocks

	if mined.Verify() != nil {
		t.Fatal("Mined block should be valid!")
	}

	m.Stop()
}
