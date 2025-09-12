package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
	"sync"
)

const MINER_REWARD = 10 // absolutely arbitrary

var ErrInvalidBlock = errors.New("block is not valid")
var ErrInsufficientBalance = errors.New("insufficient balance for transaction")

type Ledger struct {
	head     *Block
	blocks   map[[32]byte]*Block
	balances map[[32]byte]uint64 // Public keys are represented as their sha256 hash since slices can't be map indexes

	mu sync.RWMutex

	// heads []*Block // could possibly be used to handle conflicting chains, but then the balance logic would need to be head-dependant; too complex for now
}

func NewLedger(genesis Block) (*Ledger, error) {
	c := Ledger{
		blocks:   make(map[[32]byte]*Block),
		balances: make(map[[32]byte]uint64),
	}

	err := c.AddBlock(genesis)
	if err != nil {
		return nil, fmt.Errorf("making new blockchain from genesis block '%s': %w", genesis.String(), err)
	}

	return &c, nil
}

func (l *Ledger) Head() Block {
	return l.head.Clone()
}

func (l *Ledger) AddBlock(b Block) error {
	b = b.Clone()

	if b.Verify() != nil || (l.head != nil && b.PrevBlock != l.head.Hash()) {
		return ErrInvalidBlock
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	err := l.updateBalances(b)
	if err != nil {
		return err
	}

	hash := b.Hash()
	l.head = &b
	l.blocks[hash] = l.head

	return nil
}

func (l *Ledger) cloneBalances() map[[32]byte]uint64 {
	clone := make(map[[32]byte]uint64)
	maps.Copy(clone, l.balances)
	return clone
}

func (l *Ledger) updateBalances(b Block) error {
	balances := l.cloneBalances()

	for _, tx := range b.Transactions {
		if balance(balances, tx.Sender) < tx.Value {
			return ErrInsufficientBalance
		}

		decreaseBalance(balances, tx.Sender, tx.Value)
		increaseBalance(balances, tx.Receiver, tx.Value)
	}

	increaseBalance(balances, b.Miner, MINER_REWARD)

	l.balances = balances

	return nil
}

func (l *Ledger) Block(hash [32]byte) *Block {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.blocks[hash]
}

func (l *Ledger) Balance(pubkey ed25519.PublicKey) uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return balance(l.balances, pubkey)
}

func balance(balances map[[32]byte]uint64, pubkey ed25519.PublicKey) uint64 {
	hash := sha256.Sum256(pubkey)
	return balances[hash]
}

func updateBalance(balances map[[32]byte]uint64, pubkey ed25519.PublicKey, bal uint64) {
	hash := sha256.Sum256(pubkey)
	balances[hash] = bal
}

func increaseBalance(balances map[[32]byte]uint64, pubkey ed25519.PublicKey, n uint64) {
	hash := sha256.Sum256(pubkey)
	balances[hash] += n
}

func decreaseBalance(balances map[[32]byte]uint64, pubkey ed25519.PublicKey, n uint64) {
	hash := sha256.Sum256(pubkey)
	balances[hash] -= n
}
