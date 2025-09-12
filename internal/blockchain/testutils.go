package blockchain

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func AssertAddressBalance(t *testing.T, c *Ledger, a Address, expected uint64) {
	got := c.Balance(a.PublicKey())
	if got != expected {
		t.Errorf("expected balance of %d; got %d", expected, got)
	}

}

func NewTestLedger(t *testing.T, difficulty int, genesisMiner ed25519.PublicKey) (*Ledger, Block) {
	genesis := NewGenesisBlock(difficulty, genesisMiner)
	genesis.Mine()

	ledger, err := NewLedger(genesis)
	if err != nil {
		t.Fatal(err)
	}

	return ledger, genesis.Clone()
}

func GenerateTestAddress(t *testing.T) Address {
	addr, err := GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
		t.FailNow()
	}
	return addr
}
