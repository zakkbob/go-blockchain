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
			err := AddNewTestBlock(t, l, []blockchain.Transaction{tt.tx}, miner1.PublicKey())

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
