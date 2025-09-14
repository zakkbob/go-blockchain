package blockchain_test

import (
	"errors"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func TestLedgerAddBlock(t *testing.T) {
	miner := MustGenerateTestAddress(t)

	l, _ := MustCreateTestLedger(t)
	AssertAddressBalance(t, l, miner, 0)

	MustAddNewTestBlock(t, l, []blockchain.Transaction{}, miner.PublicKey())
	AssertAddressBalance(t, l, miner, 10)
}

func TestLedgerTransaction(t *testing.T) {
	miner1 := MustGenerateTestAddress(t)
	miner2 := MustGenerateTestAddress(t)

	tests := []struct {
		name                string
		tx                  blockchain.Transaction
		wantErrIs           error
		wantErrAs           any
		wantMinerOneBalance uint64
		wantMinerTwoBalance uint64
	}{
		{
			name:                "Valid transaction",
			tx:                  miner1.NewTransaction(miner2.PublicKey(), 8),
			wantErrIs:           nil,
			wantMinerOneBalance: 12,
			wantMinerTwoBalance: 8,
		},
		{
			name:                "Insufficient balance",
			tx:                  miner1.NewTransaction(miner2.PublicKey(), 12),
			wantErrIs:           blockchain.ErrInsufficientBalance,
			wantMinerOneBalance: 10,
			wantMinerTwoBalance: 0,
		},
		{
			name:                "Transaction of zero",
			tx:                  miner1.NewTransaction(miner2.PublicKey(), 0),
			wantErrAs:           &blockchain.ErrInvalidTransaction{},
			wantMinerOneBalance: 10,
			wantMinerTwoBalance: 0,
		},
		{
			name: "Unsigned transaction",
			tx: blockchain.Transaction{
				Sender:    miner1.PublicKey(),
				Receiver:  miner2.PublicKey(),
				Value:     2,
				Signature: []byte{},
			},
			wantErrAs:           &blockchain.ErrInvalidTransaction{},
			wantMinerOneBalance: 10,
			wantMinerTwoBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := MustCreateTestLedger(t)

			MustAddNewTestBlock(t, l, []blockchain.Transaction{}, miner1.PublicKey())
			_, err := AddNewTestBlock(t, l, []blockchain.Transaction{tt.tx}, miner1.PublicKey())

			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("AddBlock should return %v, not %v", tt.wantErrIs, err)
			}
			if tt.wantErrAs != nil && !errors.As(err, tt.wantErrAs) {
				t.Errorf("AddBlock should return error of type %v, not %v", tt.wantErrAs, err)
			}

			AssertAddressBalance(t, l, miner1, tt.wantMinerOneBalance)
			AssertAddressBalance(t, l, miner2, tt.wantMinerTwoBalance)

		})
	}
}

func TestLedgerNewHead(t *testing.T) {
	miner := MustGenerateTestAddress(t)

	l, genesis := MustCreateTestLedger(t)

	block1 := blockchain.NewBlock(genesis.Hash(), []blockchain.Transaction{}, 0, miner.PublicKey())
	block1.Mine()

	blockA2 := blockchain.NewBlock(block1.Hash(), []blockchain.Transaction{}, 0, miner.PublicKey())
	blockA2.Mine()
	blockA3 := blockchain.NewBlock(blockA2.Hash(), []blockchain.Transaction{}, 0, miner.PublicKey())
	blockA3.Mine()
	blockA4 := blockchain.NewBlock(blockA2.Hash(), []blockchain.Transaction{}, 0, miner.PublicKey())
	blockA4.Mine()

	blockB2 := blockchain.NewBlock(block1.Hash(), []blockchain.Transaction{}, 0, miner.PublicKey())
	blockB2.Mine()
	blockB3 := blockchain.NewBlock(blockB2.Hash(), []blockchain.Transaction{}, 0, miner.PublicKey())
	blockB3.Mine()

	MustAddTestBlock(t, l, block1)
	if l.Head().Hash() != block1.Hash() {
		t.Error("head should be block1")
	}

	MustAddTestBlock(t, l, blockA2)
	if l.Head().Hash() != blockA2.Hash() {
		t.Error("head should be blockA2")
	}

	MustAddTestBlock(t, l, blockB2)
	if l.Head().Hash() != blockA2.Hash() {
		t.Error("head should be blockA2")
	}

	MustAddTestBlock(t, l, blockB3)
	if l.Head().Hash() != blockB3.Hash() {
		t.Error("head should be blockB3")
	}

	MustAddTestBlock(t, l, blockA3)
	if l.Head().Hash() != blockB3.Hash() {
		t.Error("head should be blockB3")
	}

	MustAddTestBlock(t, l, blockA4)
	if l.Head().Hash() != blockA4.Hash() {
		t.Error("head should be blockB3")
	}

}
