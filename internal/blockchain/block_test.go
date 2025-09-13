package blockchain_test

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func TestMarshalBlock(t *testing.T) {
	addr1 := blockchain.GenerateTestAddress(t)
	addr2 := blockchain.GenerateTestAddress(t)

	tx := addr1.NewTransaction(addr2.PublicKey(), 8)

	b := blockchain.NewGenesisBlock(10)
	b.Transactions = append(b.Transactions, tx)

	js, err := json.MarshalIndent(b, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(js))
}

func TestEmptyBlockMine(t *testing.T) {
	miner, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
	}

	difficulty := 0

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

	if !(block1.VerifyHash() == nil && block2.VerifyHash() == nil) {
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

	difficulty := 0

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

	if !(block1.VerifyHash() == nil && block2.VerifyHash() == nil) {
		t.Error("Mined block should be valid")
	}
}
