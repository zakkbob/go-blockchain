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

type Blockchain struct {
	head     *Block
	blocks   map[[32]byte]*Block
	balances map[[32]byte]uint64 // Public keys are represented as their sha256 hash since slices can't be map indexes

	mu sync.RWMutex

	// heads []*Block // could possibly be used to handle conflicting chains, but then the balance logic would need to be head-dependant; too complex for now
}

func New(genesis Block) (*Blockchain, error) {
	c := Blockchain{
		blocks:   make(map[[32]byte]*Block),
		balances: make(map[[32]byte]uint64),
	}

	err := c.AddBlock(genesis)
	if err != nil {
		return nil, fmt.Errorf("making new blockchain from genesis block '%s': %w", genesis.String(), err)
	}

	return &c, nil
}

func (bc *Blockchain) Head() Block {
	return bc.head.Clone()
}

func (bc *Blockchain) AddBlock(b Block) error {
	b = b.Clone()

	if !b.Verify() || (bc.head != nil && b.PrevBlock != bc.head.Hash()) {
		return ErrInvalidBlock
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	err := bc.updateBalances(b)
	if err != nil {
		return err
	}

	hash := b.Hash()
	bc.head = &b
	bc.blocks[hash] = bc.head

	return nil
}

func (bc *Blockchain) cloneBalances() map[[32]byte]uint64 {
	clone := make(map[[32]byte]uint64)
	maps.Copy(clone, bc.balances)
	return clone
}

func (bc *Blockchain) updateBalances(b Block) error {
	balances := bc.cloneBalances()

	for _, tx := range b.Transactions {
		if balance(balances, tx.Sender) < tx.Value {
			return ErrInsufficientBalance
		}

		decreaseBalance(balances, tx.Sender, tx.Value)
		increaseBalance(balances, tx.Receiver, tx.Value)
	}

	increaseBalance(balances, b.Miner, MINER_REWARD)

	bc.balances = balances

	return nil
}

func (bc *Blockchain) Block(hash [32]byte) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.blocks[hash]
}

func (bc *Blockchain) Balance(pubkey ed25519.PublicKey) uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return balance(bc.balances, pubkey)
}

func (bc *Blockchain) updateBalance(pubkey ed25519.PublicKey, bal uint64) {
	updateBalance(bc.balances, pubkey, bal)
}

func (bc *Blockchain) increaseBalance(pubkey ed25519.PublicKey, n uint64) {
	increaseBalance(bc.balances, pubkey, n)
}

func (bc *Blockchain) decreaseBalance(pubkey ed25519.PublicKey, n uint64) {
	decreaseBalance(bc.balances, pubkey, n)
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
