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

func GenerateAddress(t *testing.T) blockchain.Address {
	addr, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
		t.FailNow()
	}
	return addr
}

func TestBlockchain(t *testing.T) {
	miner1 := GenerateAddress(t)
	miner2 := GenerateAddress(t)
	miner3 := GenerateAddress(t)

	b := blockchain.NewGenesisBlock(10, miner1.PublicKey())
	b.Mine()

	c, err := blockchain.NewBlockchain(b)
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
	}, 10, miner2.PublicKey())
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
	}, 10, miner3.PublicKey())
	b.Mine()

	err = c.AddBlock(b)
	if !errors.Is(err, blockchain.InsufficientBalance) {
		t.Errorf("AddBlock should return error InsufficientBalance, not '%v'", err)
		t.FailNow()
	}

	assertAddressBalance(t, c, miner1, 2)
	assertAddressBalance(t, c, miner2, 18)
	assertAddressBalance(t, c, miner3, 0)
}
