package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/holiman/uint256"
)

var pi, _ = uint256.FromDecimal("31415926535897932384626433832795028841971693993751058209749445923078164062862")

type Block struct {
	Difficulty   int
	PrevBlock    [32]byte
	Nonce        uint64
	Transactions []Transaction
	Timestamp    int64
	Miner        ed25519.PublicKey
	Genesis      bool

	hashLower *uint256.Int // Valid hash lower bound
	hashUpper *uint256.Int // Valid hash upper bound
}

func NewGenesisBlock(difficulty int, miner ed25519.PublicKey) Block {
	return Block{
		Difficulty:   difficulty,
		PrevBlock:    [32]byte{},
		Transactions: []Transaction{},
		Nonce:        0,
		Timestamp:    time.Now().Unix(),
		Miner:        miner,
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

	var hashLower *uint256.Int = nil
	var hashUpper *uint256.Int = nil

	if b.hashLower != nil {
		hashLower = b.hashLower.Clone()
	}
	if b.hashUpper != nil {
		hashUpper = b.hashUpper.Clone()
	}

	return Block{
		Difficulty:   b.Difficulty,
		PrevBlock:    b.PrevBlock,
		Transactions: newTxs,
		Nonce:        b.Nonce,
		Timestamp:    b.Timestamp,
		Miner:        slices.Clone(b.Miner),
		Genesis:      b.Genesis,
		hashLower:    hashLower,
		hashUpper:    hashUpper,
	}
}

func (b *Block) uint256Hash() *uint256.Int {
	hash := b.Hash()
	uint256Hash := uint256.NewInt(0)
	uint256Hash.SetBytes(hash[:])
	return uint256Hash
}

func (b *Block) String() string {
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
		b.VerifyHash(),
		b.VerifyTransactions(),
	)
}

func (b *Block) Hash() [32]byte {
	txsHash := hashTransactions(b.Transactions)

	data := b.PrevBlock[:]
	data = append(data, txsHash[:]...)
	data = binary.LittleEndian.AppendUint64(data, uint64(b.Nonce))
	data = binary.LittleEndian.AppendUint64(data, uint64(b.Timestamp))
	data = append(data, []byte(b.Miner)...)

	return sha256.Sum256(data)
}

func (b *Block) VerifyTransactions() bool {
	for _, tx := range b.Transactions {
		if !tx.Verify() {
			return false
		}
	}
	return true

}

func (b *Block) calculateHashBounds() {
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

	b.hashLower = lower
	b.hashUpper = upper
}

func (b *Block) VerifyHash() bool {
	hash := b.uint256Hash()

	return hash.Gt(b.hashLower) && hash.Lt(b.hashUpper)
}

func (b *Block) Verify() bool {
	if b.Genesis {
		return b.VerifyHash() &&
			b.VerifyTransactions()
	}

	return b.VerifyHash() &&
		b.VerifyTransactions() &&
		len(b.Transactions) > 0
}

func (b *Block) Mine() {
	if b.hashLower == nil || b.hashUpper == nil {
		b.calculateHashBounds()
	}

	for !b.VerifyHash() {
		b.Nonce += 1
	}
}
