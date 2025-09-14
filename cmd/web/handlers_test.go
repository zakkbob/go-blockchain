package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

type testLogger struct {
	t *testing.T
}

func (logger testLogger) Write(p []byte) (int, error) {
	logger.t.Log(string(p))
	return len(p), nil
}

func TestNewBlockHandler(t *testing.T) {
	addr1 := blockchain.MustGenerateTestAddress(t)
	addr2 := blockchain.MustGenerateTestAddress(t)

	ledger, genesis := blockchain.MustCreateTestLedger(t)

	app := application{
		config: config{
			debug: true,
		},
		logger: slog.New(slog.NewTextHandler(testLogger{t}, nil)),
		ledger: ledger,
	}

	srv := httptest.NewServer(app.routes())

	tx := addr1.NewTransaction(addr2.PublicKey(), 5)

	block := blockchain.NewBlock(genesis.Hash(), []blockchain.Transaction{tx}, 3, addr2.PublicKey())
	block.Mine()

	if err := block.Verify(); err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(block)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/block", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(body))
	t.Log(ledger.Head())
}
