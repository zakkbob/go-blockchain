package blockchain_test

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func assertAddressBalance(t *testing.T, c *blockchain.Blockchain, a blockchain.Address, expected uint64) {
	got := c.Balance(a.PublicKey())
	if got != expected {
		t.Errorf("expected balance of %d; got %d", expected, got)
	}

}

func generateTestAddress(t *testing.T) blockchain.Address {
	addr, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
		t.FailNow()
	}
	return addr
}

func TestBlockchain(t *testing.T) {
	miner1 := generateTestAddress(t)
	miner2 := generateTestAddress(t)
	miner3 := generateTestAddress(t)

	b := blockchain.NewGenesisBlock(2, miner1.PublicKey())
	b.Mine()

	c, err := blockchain.New(b)
	if err != nil {
		t.Errorf("NewBlockchain should not return error: %v", err)
		t.FailNow()
	}

	assertAddressBalance(t, c, miner1, 10)
	assertAddressBalance(t, c, miner2, 0)
	assertAddressBalance(t, c, miner3, 0)

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

	assertAddressBalance(t, c, miner1, 2)
	assertAddressBalance(t, c, miner2, 18)
	assertAddressBalance(t, c, miner3, 0)

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

	assertAddressBalance(t, c, miner1, 2)
	assertAddressBalance(t, c, miner2, 18)
	assertAddressBalance(t, c, miner3, 0)
}
