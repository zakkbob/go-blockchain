package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"slices"
)

type ErrInvalidTransaction struct {
	reason string
	tx     Transaction
}

func (e ErrInvalidTransaction) Error() string {
	if e.reason != "" {
		return fmt.Sprintf("invalid transaction (%s) %+v", e.reason, e.tx)
	}
	return fmt.Sprintf("invalid transaction %+v", e.tx)
}

type Transaction struct {
	Sender    ed25519.PublicKey `json:"sender"`
	Receiver  ed25519.PublicKey `json:"receiver"`
	Value     uint64            `json:"value"`
	Signature []byte            `json:"signature"`
}

func (tx Transaction) String() string {
	return fmt.Sprintf("{Sender:%s Receiver:%s Value:%d Signature:%s}",
		hex.EncodeToString(tx.Sender),
		hex.EncodeToString(tx.Receiver),
		tx.Value,
		hex.EncodeToString(tx.Signature),
	)
}

func (tx *Transaction) Clone() Transaction {
	return Transaction{
		Sender:    slices.Clone(tx.Sender),
		Receiver:  slices.Clone(tx.Receiver),
		Value:     tx.Value,
		Signature: slices.Clone(tx.Signature),
	}
}

func (tx *Transaction) Verify() error {
	hash := tx.Hash()
	if !ed25519.Verify(tx.Sender, hash[:], tx.Signature) {
		return ErrInvalidTransaction{tx: *tx, reason: "invalid signature"}
	}
	if tx.Value == 0 {
		return ErrInvalidTransaction{tx: *tx, reason: "value is 0"}
	}
	return nil
}

func (tx *Transaction) Hash() [32]byte {
	return hashTransaction(tx.Sender, tx.Receiver, tx.Value)
}

func hashTransaction(sender ed25519.PublicKey, receiver ed25519.PublicKey, value uint64) [32]byte {
	data := []byte(sender)[:]
	data = append(data, []byte(receiver)[:]...)
	data = binary.LittleEndian.AppendUint64(data, uint64(value))
	return sha256.Sum256(data)
}

func HashTransactions(txs []Transaction) [32]byte {
	var hashData []byte

	for _, tx := range txs {
		hash := tx.Hash()
		hashData = append(hashData, hash[:]...)
	}

	return sha256.Sum256(hashData)
}
