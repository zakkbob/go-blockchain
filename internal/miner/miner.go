package miner

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"sync"
	"sync/atomic"

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

	stopped atomic.Bool

	nonces         chan uint64
	correctNonces  chan uint64
	stopWorking    chan struct{}
	stopGenerating chan struct{}
	wg             sync.WaitGroup
}

func NewMiner(pubkey ed25519.PublicKey, workers int) *Miner {
	m := &Miner{
		pubkey:         pubkey,
		MinedBlocks:    make(chan *blockchain.Block),
		nonces:         make(chan uint64),
		correctNonces:  make(chan uint64),
		stopGenerating: make(chan struct{}),
		stopWorking:    make(chan struct{}),
	}

	for range workers {
		m.wg.Go(m.work)
	}

	return m
}

func (m *Miner) SetTargetBlock(b blockchain.Block) {
	close(m.stopGenerating)

	b = b.Clone()
	m.block = &b
	m.difficulty = m.block.Difficulty

	txsHash := blockchain.HashTransactions(b.Transactions)

	data := b.PrevBlock[:]
	data = append(data, txsHash[:]...)
	data = append(data, []byte(b.Miner)...)
	data = binary.LittleEndian.AppendUint64(data, uint64(b.Timestamp))

	m.partialBlockData = data

	m.stopGenerating = make(chan struct{})
	go m.generateNonces()
}

func (m *Miner) Stop() {
	if m.stopped.Load() {
		panic("stopping stopped miner")
	}

	stopDraining := make(chan struct{})

	go func() {
		for {
			select {
			case _ = <-m.correctNonces:
			case <-stopDraining:
				return
			}
		}
	}()

	m.stopped.Store(true)
	close(m.stopWorking)
	m.wg.Wait()
	close(m.stopGenerating)
	close(m.nonces)
	close(m.correctNonces)
	close(stopDraining)
}

func (m *Miner) generateNonces() {
	var nonce uint64
	var n uint64
	for {
		select {
		case n = <-m.correctNonces:
			m.block.Nonce = n
			m.MinedBlocks <- m.block
			m.block = nil
			return
		case <-m.stopGenerating:
			return
		case m.nonces <- nonce:
			nonce++
		}
	}
}

func (m *Miner) work() {
	var nonce uint64
	var data []byte
	var hash [32]byte

	for {
		select {
		case nonce = <-m.nonces:
			data = binary.LittleEndian.AppendUint64(m.partialBlockData, uint64(nonce))
			hash = sha256.Sum256(data)

			if m.checkHash(hash) {
				m.correctNonces <- nonce
			}
		case <-m.stopWorking:
			return
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
