package blockchain_test

import (
	"errors"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func TestLedger(t *testing.T) {
	miner1 := blockchain.GenerateTestAddress(t)
	miner2 := blockchain.GenerateTestAddress(t)
	miner3 := blockchain.GenerateTestAddress(t)

	blockchain.NewTestLedger(t, 2)
	c, err := blockchain.NewLedger(2)
	if err != nil {
		t.Fatalf("New should not return error: %v", err)
	}

	blockchain.AssertAddressBalance(t, c, miner1, 0)
	blockchain.AssertAddressBalance(t, c, miner2, 0)
	blockchain.AssertAddressBalance(t, c, miner3, 0)

	head := c.Head()
	b := blockchain.NewBlock(head.Hash(), []blockchain.Transaction{}, 2, miner2.PublicKey())
	b.Mine()

	err = c.AddBlock(b)
	if err != nil {
		t.Fatalf("AddBlock should not return error: %v", err)
	}

	blockchain.AssertAddressBalance(t, c, miner1, 0)
	blockchain.AssertAddressBalance(t, c, miner2, 10)
	blockchain.AssertAddressBalance(t, c, miner3, 0)

	head = c.Head()
	b = blockchain.NewBlock(head.Hash(), []blockchain.Transaction{
		miner2.NewTransaction(miner3.PublicKey(), 18),
	}, 2, miner3.PublicKey())
	b.Mine()

	err = c.AddBlock(b)
	if !errors.Is(err, blockchain.ErrInsufficientBalance) {
		t.Fatalf("AddBlock should return error InsufficientBalance, not '%v'", err)
	}

	blockchain.AssertAddressBalance(t, c, miner1, 0)
	blockchain.AssertAddressBalance(t, c, miner2, 10)
	blockchain.AssertAddressBalance(t, c, miner3, 0)
}
