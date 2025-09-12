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

	b := blockchain.NewGenesisBlock(2, miner1.PublicKey())
	b.Mine()

	c, err := blockchain.NewLedger(b)
	if err != nil {
		t.Errorf("New should not return error: %v", err)
		t.FailNow()
	}

	blockchain.AssertAddressBalance(t, c, miner1, 10)
	blockchain.AssertAddressBalance(t, c, miner2, 0)
	blockchain.AssertAddressBalance(t, c, miner3, 0)

	head := c.Head()
	b = blockchain.NewBlock(head.Hash(), []blockchain.Transaction{
		miner1.NewTransaction(miner2.PublicKey(), 8),
	}, 2, miner2.PublicKey())
	b.Mine()

	err = c.AddBlock(b)
	if err != nil {
		t.Errorf("AddBlock should not return error: %v", err)
		t.FailNow()
	}

	blockchain.AssertAddressBalance(t, c, miner1, 2)
	blockchain.AssertAddressBalance(t, c, miner2, 18)
	blockchain.AssertAddressBalance(t, c, miner3, 0)

	head = c.Head()
	b = blockchain.NewBlock(head.Hash(), []blockchain.Transaction{
		miner2.NewTransaction(miner3.PublicKey(), 20),
	}, 2, miner3.PublicKey())
	b.Mine()

	err = c.AddBlock(b)
	if !errors.Is(err, blockchain.ErrInsufficientBalance) {
		t.Errorf("AddBlock should return error InsufficientBalance, not '%v'", err)
		t.FailNow()
	}

	blockchain.AssertAddressBalance(t, c, miner1, 2)
	blockchain.AssertAddressBalance(t, c, miner2, 18)
	blockchain.AssertAddressBalance(t, c, miner3, 0)
}
