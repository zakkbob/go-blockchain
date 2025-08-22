package blockchain

import "crypto/sha256"

type Transaction struct{}

func (tx *Transaction) Hash() [32]byte {
	return sha256.Sum256([]byte{})
}

func HashTransactions(txs []Transaction) [32]byte {
	var hashData []byte

	for _, tx := range txs {
		hash := tx.Hash()
		hashData = append(hashData, hash[:]...)
	}

	return sha256.Sum256(hashData)
}
