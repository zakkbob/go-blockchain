package blockchain

import (
	"crypto/rand"
	"testing"
)

func AssertAddressBalance(t *testing.T, c *Ledger, a Address, expected uint64) {
	t.Helper()
	got := c.Balance(a.PublicKey())
	if got != expected {
		t.Errorf("expected balance of %d; got %d", expected, got)
	}

}

func NewTestLedger(t *testing.T, difficulty int) (*Ledger, Block) {
	t.Helper()
	ledger, err := NewLedger(difficulty)
	if err != nil {
		t.Fatal(err)
	}

	return ledger, ledger.Head()
}

func GenerateTestAddress(t *testing.T) Address {
	t.Helper()
	addr, err := GenerateAddress(rand.Reader)
	if err != nil {
		t.Errorf("GenerateAddress should not return an error: %v", err)
		t.FailNow()
	}
	return addr
}
