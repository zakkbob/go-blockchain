package miner_test

import (
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/miner"
)

func TestMiner(t *testing.T) {
	miner1 := blockchain.MustGenerateTestAddress(t)

	b := blockchain.NewGenesisBlock(5)

	m := miner.NewMiner(miner1.PublicKey())
	m.Mine(b)
	mined := <-m.MinedBlocks

	if mined.Verify() != nil {
		t.Fatal("Mined block should be valid!")
	}

	m.Stop()
}
