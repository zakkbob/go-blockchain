package miner

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/holiman/uint256"
	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

var pi, _ = uint256.FromDecimal("31415926535897932384626433832795028841971693993751058209749445923078164062862")

type Miner struct {
	MinedBlocks chan *blockchain.Block

	pubkey           ed25519.PublicKey
	block            *blockchain.Block
	partialBlockData []byte
	difficulty       int

	sendCorrectNonceOnce sync.Once
	stopOnce             sync.Once

	stopWorking chan struct{}
	wg          sync.WaitGroup
}

func NewMiner(pubkey ed25519.PublicKey) *Miner {
	m := &Miner{
		pubkey:      pubkey,
		MinedBlocks: make(chan *blockchain.Block),
		stopWorking: make(chan struct{}),
	}

	return m
}

// Starts mining a block, using one worker
// Can be called while already mining
func (m *Miner) Mine(b blockchain.Block) {
	fmt.Println("Starting new mining work")
	m.Stop()

	b = b.Clone()
	m.block = &b
	m.difficulty = m.block.Difficulty

	txsHash := blockchain.HashTransactions(b.Transactions)

	data := b.PrevBlock[:]
	data = append(data, txsHash[:]...)
	data = append(data, []byte(b.Miner)...)
	data = binary.LittleEndian.AppendUint64(data, uint64(b.Timestamp))

	m.partialBlockData = data

	m.stopWorking = make(chan struct{})
	m.sendCorrectNonceOnce = sync.Once{}
	m.stopOnce = sync.Once{}

	m.wg.Go(func() {
		m.work()
	})
}

// Stops mining, can be called multiple times safely
func (m *Miner) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopWorking)
		m.wg.Wait()
	})
}

func (m *Miner) processCorrectNonce(n uint64) {
	m.sendCorrectNonceOnce.Do(func() {
		m.block.Nonce = n
		m.MinedBlocks <- m.block
		m.partialBlockData = nil
		m.block = nil
	})
}

func (m *Miner) work() {
	var nonce uint64
	var data []byte
	var hash [32]byte

	for {
		select {
		case <-m.stopWorking:
			return
		default:
			data = binary.LittleEndian.AppendUint64(m.partialBlockData, uint64(nonce))
			hash = sha256.Sum256(data)

			if m.checkHash(hash) {
				m.processCorrectNonce(nonce)
			}

			nonce++
		}
	}

}

func (m *Miner) checkHash(hash [32]byte) bool {
	digits := 77 - m.difficulty/3
	divisor := math.Pow(2, float64(m.difficulty%3))

	lower := pi.Clone()
	div := uint256.NewInt(10)
	exp := uint256.NewInt(uint64(digits))

	div.Exp(div, exp)
	lower.Div(lower, div)
	lower.Mul(lower, div)

	div.Div(div, uint256.NewInt(uint64(divisor)))

	upper := lower.Clone()
	upper.Add(upper, div)

	uint256Hash := uint256.NewInt(0)
	uint256Hash.SetBytes(hash[:])

	return uint256Hash.Gt(lower) && uint256Hash.Lt(upper)
}
