package blockchain_test

import (
	"crypto/rand"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func TestEmptyBlockMine(t *testing.T) {
	miner, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
	}

	difficulty := 10

	block1 := blockchain.NewBlock(
		[32]byte{},
		[]blockchain.Transaction{},
		difficulty,
		miner.PublicKey(),
	)

	block1.Mine()

	block2 := blockchain.NewBlock(
		block1.Hash(),
		[]blockchain.Transaction{},
		difficulty,
		miner.PublicKey(),
	)
	block2.Mine()

	t.Log(block1.String())
	t.Log(block2.String())

	if !(block1.ValidHash() && block2.ValidHash()) {
		t.Error("Mined block should be valid")
	}
}

func TestTransactionBlockMine(t *testing.T) {
	miner, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
	}

	sender, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
	}

	receiver, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
	}

	tx := sender.NewTransaction(receiver.PublicKey(), 1)

	difficulty := 20

	block1 := blockchain.NewBlock(
		[32]byte{},
		[]blockchain.Transaction{},
		difficulty,
		miner.PublicKey(),
	)

	block1.Mine()

	block2 := blockchain.NewBlock(
		block1.Hash(),
		[]blockchain.Transaction{tx},
		difficulty,
		miner.PublicKey(),
	)
	block2.Mine()

	t.Log(block1.String())
	t.Log(block2.String())

	if !(block1.ValidHash() && block2.ValidHash()) {
		t.Error("Mined block should be valid")
	}
}
