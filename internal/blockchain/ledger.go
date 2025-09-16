package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
	"sync"

	"github.com/holiman/uint256"
)

const MINER_REWARD = 10 // absolutely arbitrary

var (
	ErrInsufficientBalance = errors.New("insufficient balance for transaction")
)

type ErrPrevBlockNotFound struct {
	hash [32]byte
}

func (e ErrPrevBlockNotFound) Error() string {
	uint256Hash := uint256.NewInt(0)
	uint256Hash.SetBytes(e.hash[:])
	return fmt.Sprintf("previous block with hash %s could not be found", uint256Hash.Dec())
}

type balances struct {
	balances map[[32]byte]uint64 // indexed by public key hash
}

func (b *balances) Clone() balances {
	return balances{
		balances: maps.Clone(b.balances),
	}
}

func (b *balances) Get(pubkey ed25519.PublicKey) uint64 {
	hash := sha256.Sum256(pubkey)
	return b.balances[hash]
}

func (b *balances) Set(pubkey ed25519.PublicKey, bal uint64) {
	hash := sha256.Sum256(pubkey)
	b.balances[hash] = bal
}

func (b *balances) Increase(pubkey ed25519.PublicKey, n uint64) {
	hash := sha256.Sum256(pubkey)
	b.balances[hash] += n
}

func (b *balances) Decrease(pubkey ed25519.PublicKey, n uint64) {
	hash := sha256.Sum256(pubkey)
	b.balances[hash] -= n
}

type head struct {
	block    *Block
	length   int
	balances balances
}

func (h *head) ConstructNextBlock(txPool map[[32]byte]Transaction, miner ed25519.PublicKey) Block {
	balances := h.balances.Clone()

	var txs []Transaction

	for _, tx := range txPool {
		if tx.Value == 0 {
			panic("oh no")
		}

		if balances.Get(tx.Sender) < tx.Value {
			continue
		}

		balances.Decrease(tx.Sender, tx.Value)
		balances.Increase(tx.Receiver, tx.Value)

		txs = append(txs, tx)
	}

	return NewBlock(h.block.Hash(), txs, h.block.Difficulty, miner)
}

func (h *head) Update(b *Block) error {
	if b.PrevBlock != h.block.Hash() {
		panic("what in the heck")
	}

	balances := h.balances.Clone()

	for _, tx := range b.Transactions {
		if tx.Value == 0 {
			return ErrInvalidTransaction{tx: tx}
		}

		if balances.Get(tx.Sender) < tx.Value {
			return ErrInsufficientBalance
		}

		balances.Decrease(tx.Sender, tx.Value)
		balances.Increase(tx.Receiver, tx.Value)
	}

	balances.Increase(b.Miner, MINER_REWARD)

	h.balances = balances
	h.block = b
	h.length++

	return nil
}

type Ledger struct {
	blocks map[[32]byte]*Block // All known, verified blocks
	heads  []*head             // All possible heads of chains from the known blocks
	head   *head               // The longest head (chain with most work)

	mu sync.RWMutex
}

func NewLedger(difficulty int) (*Ledger, error) {
	genesis := NewGenesisBlock(difficulty)
	genesis.Mine()

	balances := balances{
		balances: map[[32]byte]uint64{},
	}

	h := &head{
		block:    &genesis,
		length:   1,
		balances: balances,
	}

	blocks := map[[32]byte]*Block{}
	blocks[genesis.Hash()] = &genesis

	c := Ledger{
		blocks: blocks,
		heads:  []*head{h},
		head:   h,
	}

	return &c, nil
}

func (l *Ledger) Length() int {
	return l.head.length
}

func (l *Ledger) ConstructNextBlock(txPool map[[32]byte]Transaction, miner ed25519.PublicKey) Block {
	return l.head.ConstructNextBlock(txPool, miner)
}

func (l *Ledger) getHead(hash [32]byte) (*head, bool) {
	for _, h := range l.heads {
		if h.block.Hash() == hash {
			return h, true
		}
	}
	return nil, false
}

func (l *Ledger) getChain(hash [32]byte) []*Block {
	chain := []*Block{}

	for {
		block, ok := l.blocks[hash]
		if !ok {
			break
		}

		chain = append(chain, block)
		hash = block.PrevBlock
	}

	if hash != [32]byte{} {
		panic("oh no")
	}

	return chain
}

func (l *Ledger) headFromBlock(hash [32]byte) *head {
	c := l.getChain(hash)
	genesis := c[len(c)-1]

	h := head{
		block:  genesis,
		length: 1,
		balances: balances{
			balances: map[[32]byte]uint64{},
		},
	}

	for i := len(c) - 2; i >= 0; i-- {
		h.Update(c[i])
	}

	return &h
}

func (l *Ledger) AddBlock(b Block) error {
	b = b.Clone()

	if err := b.Verify(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.blocks[b.PrevBlock]; !ok {
		return ErrPrevBlockNotFound{hash: b.PrevBlock}
	}

	h, ok := l.getHead(b.PrevBlock)
	if !ok {
		h = l.headFromBlock(b.PrevBlock)
	}

	err := h.Update(&b)
	if err != nil {
		return err
	}
	if h.length > l.head.length {
		l.head = h
	}

	l.blocks[b.Hash()] = &b

	return nil
}

func (l *Ledger) Head() *Block {
	l.mu.RLock()
	defer l.mu.RUnlock()
	b := l.head.block.Clone()
	return &b
}

func (l *Ledger) Balance(pubkey ed25519.PublicKey) uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.head.balances.Get(pubkey)
}
