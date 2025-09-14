package blockchain

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func MustAddTestBlock(t *testing.T, l *Ledger, b Block) {
	t.Helper()
	b.Mine()
	err := l.AddBlock(b)
	if err != nil {
		t.Fatalf("AddBlock should not return error: %v", err)
	}
}

func AddNewTestBlock(t *testing.T, l *Ledger, txs []Transaction, miner ed25519.PublicKey) (*Block, error) {
	t.Helper()
	head := l.Head()
	b := NewBlock(head.Hash(), txs, 0, miner)
	b.Mine()

	return &b, l.AddBlock(b)
}

func MustAddNewTestBlock(t *testing.T, l *Ledger, txs []Transaction, miner ed25519.PublicKey) *Block {
	t.Helper()
	b, err := AddNewTestBlock(t, l, txs, miner)
	if err != nil {
		t.Fatalf("AddBlock should not return error: %v", err)
	}
	return b
}

func AssertAddressBalance(t *testing.T, l *Ledger, a Address, expected uint64) {
	t.Helper()
	got := l.Balance(a.PublicKey())
	if got != expected {
		t.Errorf("expected balance of %d; got %d", expected, got)
	}

}

func MustCreateTestLedger(t *testing.T) (*Ledger, *Block) {
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
