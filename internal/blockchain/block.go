package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/holiman/uint256"
	"github.com/zakkbob/go-blockchain/internal/bits"
)

type Block struct {
	difficulty   int
	prevBlock    [32]byte
	nonce        *uint256.Int
	transactions []Transaction
	timestamp    int64
}

func NewBlock(prevBlock [32]byte, txs []Transaction, difficulty int) Block {
	return Block{
		difficulty:   difficulty,
		prevBlock:    prevBlock,
		transactions: txs,
		nonce:        uint256.NewInt(0),
		timestamp:    time.Now().Unix(),
	}
}

func (b *Block) String() string {
	hash := b.Hash()
	return fmt.Sprintf(
		"Hash: %s; Previous Block: %s; Difficulty: %d; Transactions: %d; Nonce: %s; Timestamp: %d; Valid: %t;",
		hex.EncodeToString(hash[:]),
		hex.EncodeToString(b.prevBlock[:]),
		b.difficulty,
		len(b.transactions),
		b.nonce.String(),
		b.timestamp,
		b.Valid(),
	)
}

func (b *Block) Timestamp() int64 {
	return b.timestamp
}

func (b *Block) Hash() [32]byte {
	nonceBytes := b.nonce.Bytes32()
	txsHash := HashTransactions(b.transactions)

	data := b.prevBlock[:]
	data = append(data, txsHash[:]...)
	data = append(data, nonceBytes[:]...)
	data = binary.LittleEndian.AppendUint64(data, uint64(b.timestamp))

	return sha256.Sum256(data)
}

func (b *Block) Nonce() *uint256.Int {
	return b.nonce
}

func (b *Block) Valid() bool {
	hash := b.Hash()
	zeros := bits.LeadingZerosBytes(hash[:])

	return zeros >= b.difficulty
}

func (b *Block) Mine() {
	one := uint256.NewInt(1)
	for !b.Valid() {
		b.nonce.Add(b.nonce, one)
	}
}
