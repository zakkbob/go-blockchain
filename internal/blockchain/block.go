package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/holiman/uint256"
)

var (
	ErrHashOutOfBounds = errors.New("hash is not within required boundaries")
)

var pi, _ = uint256.FromDecimal("31415926535897932384626433832795028841971693993751058209749445923078164062862")

type Block struct {
	Difficulty   int               `json:"difficulty"`
	PrevBlock    [32]byte          `json:"previous_block"`
	Nonce        uint64            `json:"nonce"`
	Transactions []Transaction     `json:"transactions"`
	Timestamp    int64             `json:"timestamp"`
	Miner        ed25519.PublicKey `json:"miner"`
	Genesis      bool              `json:"genesis"`
}

func NewGenesisBlock(difficulty int) Block {
	return Block{
		Difficulty:   difficulty,
		PrevBlock:    [32]byte{},
		Transactions: []Transaction{},
		Nonce:        0,
		Timestamp:    time.Now().Unix(),
		Miner:        ed25519.PublicKey{},
		Genesis:      true,
	}
}

func NewBlock(prevBlock [32]byte, txs []Transaction, difficulty int, miner ed25519.PublicKey) Block {
	return Block{
		Difficulty:   difficulty,
		PrevBlock:    prevBlock,
		Transactions: txs,
		Nonce:        0,
		Timestamp:    time.Now().Unix(),
		Miner:        miner,
		Genesis:      false,
	}
}

// This might be avoided by creating a new struct MinedBlock which is immutable and returned by Mine()
func (b *Block) Clone() Block {
	newTxs := make([]Transaction, len(b.Transactions))

	for i, tx := range b.Transactions {
		newTxs[i] = tx.Clone()
	}

	return Block{
		Difficulty:   b.Difficulty,
		PrevBlock:    b.PrevBlock,
		Transactions: newTxs,
		Nonce:        b.Nonce,
		Timestamp:    b.Timestamp,
		Miner:        slices.Clone(b.Miner),
		Genesis:      b.Genesis,
	}
}

func (b *Block) uint256Hash() *uint256.Int {
	hash := b.Hash()
	uint256Hash := uint256.NewInt(0)
	uint256Hash.SetBytes(hash[:])
	return uint256Hash
}

func (b Block) String() string {
	hash := b.uint256Hash()
	return fmt.Sprintf(
		"Hash: %s; Previous Block: %s; Difficulty: %d; Transactions: %d; Nonce: %d; Timestamp: %d; Mined By: %s; Genesis: %t; Hash Satisfies Difficulty: %t; Verified Transactions: %t",
		hash.Dec(),
		hex.EncodeToString(b.PrevBlock[:]),
		b.Difficulty,
		len(b.Transactions),
		b.Nonce,
		b.Timestamp,
		hex.EncodeToString(b.Miner),
		b.Genesis,
		b.VerifyHash() == nil,
		b.VerifyTransactions() == nil,
	)
}

func (b *Block) Hash() [32]byte {
	txsHash := HashTransactions(b.Transactions)

	data := b.PrevBlock[:]
	data = append(data, txsHash[:]...)
	data = append(data, []byte(b.Miner)...)
	data = binary.LittleEndian.AppendUint64(data, uint64(b.Timestamp))
	data = binary.LittleEndian.AppendUint64(data, uint64(b.Nonce))

	return sha256.Sum256(data)
}

func (b *Block) VerifyTransactions() error {
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}
	return nil

}

func (b *Block) VerifyHash() error {
	hash := b.uint256Hash()

	digits := 77 - b.Difficulty/3
	divisor := math.Pow(2, float64(b.Difficulty%3))

	lower := pi.Clone()
	div := uint256.NewInt(10)
	exp := uint256.NewInt(uint64(digits))

	div.Exp(div, exp)
	lower.Div(lower, div)
	lower.Mul(lower, div)

	div.Div(div, uint256.NewInt(uint64(divisor)))

	upper := lower.Clone()
	upper.Add(upper, div)

	if !hash.Gt(lower) || !hash.Lt(upper) {
		return ErrHashOutOfBounds
	}

	return nil
}

func (b *Block) Verify() error {
	if err := b.VerifyHash(); err != nil {
		return err
	}
	if err := b.VerifyTransactions(); err != nil {
		return err
	}

	return nil
}

func (b *Block) Mine() {
	for b.VerifyHash() != nil {
		b.Nonce += 1
	}
}
