package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
)

type Transaction struct {
	sender    ed25519.PublicKey
	receiver  ed25519.PublicKey
	value     uint64
	signature []byte
}

func MakeTransaction(sender ed25519.PublicKey, receiver ed25519.PublicKey, value uint64, signature []byte) Transaction {
	return Transaction{
		sender:    sender,
		receiver:  receiver,
		value:     value,
		signature: signature,
	}
}

func (tx *Transaction) Verify() bool {
	hash := tx.Hash()
	return ed25519.Verify(tx.sender, hash[:], tx.signature)
}

func HashTransaction(sender ed25519.PublicKey, receiver ed25519.PublicKey, value uint64) [32]byte{
	data := []byte(sender)[:]
	data = append(data, []byte(receiver)[:]...)
	data = binary.LittleEndian.AppendUint64(data, uint64(value))
	return sha256.Sum256(data)

}

func (tx *Transaction) Hash() [32]byte {
	return HashTransaction(tx.sender, tx.receiver, tx.value)
}

func HashTransactions(txs []Transaction) [32]byte {
	var hashData []byte

	for _, tx := range txs {
		hash := tx.Hash()
		hashData = append(hashData, hash[:]...)
	}

	return sha256.Sum256(hashData)
}
