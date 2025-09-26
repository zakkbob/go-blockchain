package main

import (
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
	"github.com/zakkbob/go-blockchain/internal/txpool"
)

func TestNewTransactionHandler(t *testing.T) {
	addr1 := blockchain.MustGenerateTestAddress(t)
	addr2 := blockchain.MustGenerateTestAddress(t)

	tests := []struct {
		name         string
		tx           blockchain.Transaction
		expectedSize int
	}{
		{
			name:         "valid transaction",
			tx:           addr1.NewTransaction(addr2.PublicKey(), 10),
			expectedSize: 1,
		},
		{
			name:         "transaction with 0 value",
			tx:           addr1.NewTransaction(addr2.PublicKey(), 0),
			expectedSize: 0,
		},
		{
			name: "unsigned transaction",
			tx: blockchain.Transaction{
				Sender:    addr1.PublicKey(),
				Receiver:  addr2.PublicKey(),
				Value:     5,
				Signature: []byte{},
			},
			expectedSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := application{
				logger: CreateTestLogger(t),
				config: CreateTestConfig(t),
				txpool: txpool.Pool{},
			}

			msg := gossip.CreateReceivedMessage(t, msgNewTransaction, "test :D", tt.tx)
			app.newTransactionHandler(msg)

			if app.txpool.Size() != tt.expectedSize {
				t.Fatalf("Transaction pool size should be %d", tt.expectedSize)
			}

		})

	}
}

func TestNewBlockHandler(t *testing.T) {
	addr1 := blockchain.MustGenerateTestAddress(t)
	addr2 := blockchain.MustGenerateTestAddress(t)

	ledger, genesis := blockchain.MustCreateTestLedger(t)

	app := application{
		logger: CreateTestLogger(t),
		config: CreateTestConfig(t),
		ledger: ledger,
	}

	block := blockchain.NewBlock(genesis.Hash(), []blockchain.Transaction{}, 3, addr1.PublicKey())
	block.Mine()

	tx := addr1.NewTransaction(addr2.PublicKey(), 5)
	block2 := blockchain.NewBlock(block.Hash(), []blockchain.Transaction{tx}, 3, addr2.PublicKey())
	block2.Mine()

	msg1 := gossip.CreateReceivedMessage(t, msgNewBlock, "test :D", block)
	msg2 := gossip.CreateReceivedMessage(t, msgNewBlock, "test :D", block2)

	app.newBlockHandler(msg1)
	app.newBlockHandler(msg2)

	if ledger.Length() != 3 {
		t.Fatal("Ermm, blocks should've been added!")
	}

}
