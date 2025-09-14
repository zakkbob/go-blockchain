package blockchain

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func AddNewTestBlock(t *testing.T, ledger *Ledger, txs []Transaction, miner ed25519.PublicKey) error {
	head := ledger.Head()
	b := NewBlock(head.Hash(), txs, 0, miner)
	b.Mine()

	return ledger.AddBlock(b)
}

func MustAddNewTestBlock(t *testing.T, ledger *Ledger, txs []Transaction, miner ed25519.PublicKey) {
	err := AddNewTestBlock(t, ledger, txs, miner)
	if err != nil {
		t.Fatalf("AddBlock should not return error: %v", err)
	}
}

func AssertAddressBalance(t *testing.T, c *Ledger, a Address, expected uint64) {
	t.Helper()
	got := c.Balance(a.PublicKey())
	if got != expected {
		t.Errorf("expected balance of %d; got %d", expected, got)
	}

}

func MustCreateTestLedger(t *testing.T) (*Ledger, Block) {
	t.Helper()
	ledger, err := NewLedger(0)
	if err != nil {
		t.Fatal(err)
	}

	return ledger, ledger.Head()
}

func MustGenerateTestAddress(t *testing.T) Address {
	t.Helper()
	addr, err := GenerateAddress(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateAddress should not return an error: %v", err)
	}
	return addr
}
