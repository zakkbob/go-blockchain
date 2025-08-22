package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"slices"
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
	miner        ed25519.PublicKey
	genesis      bool
}

func MakeBlock(difficulty int, prevBlock [32]byte, nonce *uint256.Int, transactions []Transaction, timestamp int64, miner ed25519.PublicKey) Block {
	return Block{
		difficulty:   difficulty,
		prevBlock:    prevBlock,
		nonce:        nonce,
		transactions: transactions,
		timestamp:    timestamp,
		miner:        miner,
		genesis:      false,
	}
}

func NewGenesisBlock(difficulty int, miner ed25519.PublicKey) Block {
	return Block{
		difficulty:   difficulty,
		prevBlock:    [32]byte{},
		transactions: []Transaction{},
		nonce:        uint256.NewInt(0),
		timestamp:    time.Now().Unix(),
		miner:        miner,
		genesis:      true,
	}
}

func NewBlock(prevBlock [32]byte, txs []Transaction, difficulty int, miner ed25519.PublicKey) Block {
	return Block{
		difficulty:   difficulty,
		prevBlock:    prevBlock,
		transactions: txs,
		nonce:        uint256.NewInt(0),
		timestamp:    time.Now().Unix(),
		miner:        miner,
		genesis:      false,
	}
}

func (b *Block) Clone() Block {
	newTxs := make([]Transaction, len(b.transactions))

	for i, tx := range b.transactions {
		newTxs[i] = tx.Clone()
	}

	return Block {
		difficulty: b.difficulty,
		prevBlock: b.prevBlock,
		transactions: newTxs,
		nonce: b.nonce,
		timestamp: b.timestamp,
		miner: slices.Clone(b.miner),
		genesis: b.genesis,
	}
}

func (b *Block) String() string {
	hash := b.Hash()
	return fmt.Sprintf(
		"Hash: %s; Previous Block: %s; Difficulty: %d; Transactions: %d; Nonce: %s; Timestamp: %d; Mined By: %s; Genesis: %t; Hash Satisfies Difficulty: %t; Verified Transactions: %t",
		hex.EncodeToString(hash[:]),
		hex.EncodeToString(b.prevBlock[:]),
		b.difficulty,
		len(b.transactions),
		b.nonce.String(),
		b.timestamp,
		hex.EncodeToString(b.miner),
		b.genesis,
		b.ValidHash(),
		b.VerifyTransactions(),
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
	data = append(data, []byte(b.miner)...)

	return sha256.Sum256(data)
}

func (b *Block) Nonce() *uint256.Int {
	return b.nonce
}

func (b *Block) VerifyTransactions() bool {
	for _, tx := range b.transactions {
		if !tx.Verify() {
			return false
		}
	}
	return true

}

func (b *Block) ValidHash() bool {
	hash := b.Hash()
	zeros := bits.LeadingZerosBytes(hash[:])

	return zeros >= b.difficulty
}

func (b *Block) ValidBlock() bool {
	if b.genesis {
		return b.ValidHash() &&
			b.VerifyTransactions()
	}

	return b.ValidHash() &&
		b.VerifyTransactions() &&
		len(b.transactions) > 0
}

func (b *Block) Mine() {
	one := uint256.NewInt(1)
	for !b.ValidHash() {
		b.nonce.Add(b.nonce, one)
	}
}
