package miner

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/holiman/uint256"
	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

var pi, _ = uint256.FromDecimal("31415926535897932384626433832795028841971693993751058209749445923078164062862")

type Miner struct {
	pubkey           ed25519.PublicKey
	ledger           *blockchain.Ledger
	txPool           map[[32]byte]blockchain.Transaction
	block            *blockchain.Block
	partialBlockData []byte

	nonceChan        chan uint64
	correctNonceChan chan uint64
	cancelChan       chan bool
}

func NewMiner(l *blockchain.Ledger, pubkey ed25519.PublicKey) Miner {
	return Miner{
		ledger:           l,
		txPool:           map[[32]byte]blockchain.Transaction{},
		pubkey:           pubkey,
		nonceChan:        make(chan uint64),
		correctNonceChan: make(chan uint64),
		cancelChan:       make(chan bool),
	}
}

func (m *Miner) AddBlock(b blockchain.Block) error {
	oldHead := m.ledger.Head().Hash()

	err := m.ledger.AddBlock(b)
	if err != nil {
		return err
	}

	if oldHead != m.ledger.Head().Hash() {
		//m.updateBlock()
	}
	return nil
}

func (m *Miner) AddTransaction(tx blockchain.Transaction) error {
	if err := tx.Verify(); err != nil {
		return err
	}
	m.txPool[tx.Hash()] = tx
	return nil
}

func (m *Miner) Mine(workers int) {
	go func() {
		var nonce uint64
		var n uint64
		for {
			select {
			case n = <-m.correctNonceChan:
				m.block.Nonce = n
				err := m.ledger.AddBlock(*m.block)
				if err != nil {
					fmt.Print(err)
				}
				m.updateBlock()

			case _ = <-m.cancelChan:
				close(m.cancelChan)
				close(m.nonceChan)
				return
			case m.nonceChan <- nonce:
				nonce++
			}
		}
	}()

	m.updateBlock()

	for range workers {
		go func() {
			var nonce uint64
			var data []byte
			var hash [32]byte

			for {
				nonce = <-m.nonceChan
				data = binary.LittleEndian.AppendUint64(m.partialBlockData, uint64(nonce))
				hash = sha256.Sum256(data)

				if m.checkHash(hash) {
					m.correctNonceChan <- nonce
				}
			}
		}()
	}
}

func (m *Miner) updateBlock() {
	b := m.ledger.ConstructNextBlock(m.txPool, m.pubkey)
	m.block = &b

	txsHash := blockchain.HashTransactions(b.Transactions)

	data := b.PrevBlock[:]
	data = append(data, txsHash[:]...)
	data = append(data, []byte(b.Miner)...)
	data = binary.LittleEndian.AppendUint64(data, uint64(b.Timestamp))

	m.partialBlockData = data
}

func (m *Miner) checkHash(hash [32]byte) bool {
	digits := 77 - m.block.Difficulty/3
	divisor := math.Pow(2, float64(m.block.Difficulty%3))

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
