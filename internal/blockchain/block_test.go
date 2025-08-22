package blockchain_test

import (
	"testing"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func TestBlockMine(t *testing.T) {
	difficulty := 10

	block1 := blockchain.NewBlock(
		[32]byte{},
		[]blockchain.Transaction{},
		difficulty,
	)
	block1.Mine()

	block2 := blockchain.NewBlock(
		block1.Hash(),
		[]blockchain.Transaction{},
		difficulty,
	)
	block2.Mine()

	t.Error(block1.String())
	t.Error(block2.String())
}
