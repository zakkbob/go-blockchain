package main

import (
	"log/slog"
	"testing"
	"time"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

type testLogger struct {
	t *testing.T
}

func (l testLogger) Write(p []byte) (int, error) {
	l.t.Log(string(p))
	return len(p), nil
}

func TestNewBlockHandler(t *testing.T) {
	addr1 := blockchain.MustGenerateTestAddress(t)
	addr2 := blockchain.MustGenerateTestAddress(t)

	ledger, genesis := blockchain.MustCreateTestLedger(t)

	logger := slog.NewLogLogger(slog.NewTextHandler(testLogger{t}, nil), slog.LevelError)

	app := application{
		config: config{
			debug: true,
		},
		logger: slog.New(slog.NewTextHandler(testLogger{t}, nil)),
		ledger: ledger,
	}

	node1 := gossip.Node{
		Addr:     ":0",
		ErrorLog: logger,
	}

	go func() {
		err := node1.BootstrapAndListen([]string{}, app.handler)
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Second)

	node2 := gossip.Node{
		Addr:     ":0",
		ErrorLog: logger,
	}

	go func() {
		err := node2.BootstrapAndListen([]string{node1.ListenerAddr().String()}, func(rm gossip.ReceivedMessage) { t.Fatal(rm) })
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Second)

	block := blockchain.NewBlock(genesis.Hash(), []blockchain.Transaction{}, 3, addr1.PublicKey())
	block.Mine()

	tx := addr1.NewTransaction(addr2.PublicKey(), 5)
	block2 := blockchain.NewBlock(block.Hash(), []blockchain.Transaction{tx}, 3, addr2.PublicKey())
	block2.Mine()

	if err := block.Verify(); err != nil {
		t.Fatal(err)
	}

	node2.Broadcast(gossip.Message{
		Type: "newBlock",
		Data: block2,
	})

	node2.Broadcast(gossip.Message{
		Type: "newBlock",
		Data: block,
	})

	node2.Broadcast(gossip.Message{
		Type: "newBlock",
		Data: block2,
	})

	time.Sleep(time.Second)
}
