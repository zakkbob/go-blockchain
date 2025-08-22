package blockchain

import (
	"crypto/ed25519"
	"fmt"
	"io"
)

type Address struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

func GenerateAddress(rand io.Reader) (Address, error) {
	public, private, err := ed25519.GenerateKey(rand)
	if err != nil {
		return Address{}, fmt.Errorf("generating ed25519 keypair: %w", err)
	}
	return Address{
		publicKey:  public,
		privateKey: private,
	}, nil
}

func (a *Address) PublicKey() ed25519.PublicKey {
	return a.publicKey
}

func (a *Address) NewTransaction(receiver ed25519.PublicKey, value uint64) Transaction {
	hash := HashTransaction(a.publicKey, receiver, value)
	sig := ed25519.Sign(a.privateKey, hash[:])
	tx := MakeTransaction(a.publicKey, receiver, value, sig)
	return tx
}
