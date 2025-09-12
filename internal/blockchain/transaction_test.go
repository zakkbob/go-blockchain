package blockchain_test

import (
	"encoding/json"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func TestMarshalTransaction(t *testing.T) {
	addr1 := blockchain.GenerateTestAddress(t)
	addr2 := blockchain.GenerateTestAddress(t)

	tx := addr1.NewTransaction(addr2.PublicKey(), 8)

	js, err := json.MarshalIndent(tx, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(js))
}
