package blockchain_test

import (
	"encoding/json"
	"testing"
)

func TestMarshalTransaction(t *testing.T) {
	addr1 := MustGenerateTestAddress(t)
	addr2 := MustGenerateTestAddress(t)

	tx := addr1.NewTransaction(addr2.PublicKey(), 8)

	js, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(js))
}
