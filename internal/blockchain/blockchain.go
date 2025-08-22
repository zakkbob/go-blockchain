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

var InvalidBlock = errors.New("block is not valid")
var InsufficientBalance = errors.New("insufficient balance for transaction")

type Blockchain struct {
	head     *Block
	blocks   map[[32]byte]*Block
	balances map[[32]byte]uint64 // Public keys are represented as their sha256 hash since slices can't be map indexes

	lock sync.RWMutex

	// heads []*Block // could possibly be used to handle conflicting chains, but then the balance logic would need to be head-dependant; too complex for now
}

func NewBlockchain(genesis Block) (*Blockchain, error) {
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

func (c *Blockchain) Head() Block {
	return c.head.Clone()
}

func (c *Blockchain) AddBlock(b Block) error {
	clone := b.Clone()

	if !clone.ValidBlock() {
		return InvalidBlock
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	err := c.updateBalances(clone)
	if err != nil {
		return err
	}

	hash := clone.Hash()
	c.head = &clone
	c.blocks[hash] = c.head

	return nil
}

func (c *Blockchain) cloneBalances() map[[32]byte]uint64 {
	clone := make(map[[32]byte]uint64)
	maps.Copy(clone, c.balances)
	return clone
}

func (c *Blockchain) updateBalances(b Block) error {
	balances := c.cloneBalances()

	for _, tx := range b.transactions {
		if balance(balances, tx.sender) < tx.value {
			return InsufficientBalance
		}

		decreaseBalance(balances, tx.sender, tx.value)
		increaseBalance(balances, tx.receiver, tx.value)
	}

	increaseBalance(balances, b.miner, MINER_REWARD)

	c.balances = balances

	return nil
}

func (c *Blockchain) Block(hash [32]byte) *Block {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.blocks[hash]
}

func (c *Blockchain) Balance(pubkey ed25519.PublicKey) uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return balance(c.balances, pubkey)
}

func (c *Blockchain) updateBalance(pubkey ed25519.PublicKey, bal uint64) {
	updateBalance(c.balances, pubkey, bal)
}

func (c *Blockchain) increaseBalance(pubkey ed25519.PublicKey, n uint64) {
	increaseBalance(c.balances, pubkey, n)
}

func (c *Blockchain) decreaseBalance(pubkey ed25519.PublicKey, n uint64) {
	decreaseBalance(c.balances, pubkey, n)
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
