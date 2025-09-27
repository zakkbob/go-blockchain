package main

import (
	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) updateMiningTarget() {
	b := app.constructNextBlock()
	app.miner.SetTargetBlock(b)
}

func (app *application) constructNextBlock() blockchain.Block {
	var (
		prevHash   = app.ledger.HeadHash()
		difficulty = app.ledger.CalculateFutureDifficulty()
		balances   = app.ledger.Balances()
		candidates = app.txpool.Get(100)
		txs        = make([]blockchain.Transaction, 0, 100)
	)

	for _, tx := range candidates {
		if err := tx.Verify(); err != nil { //sanity check
			panic("oh no")
		}

		if balances.Get(tx.Sender) < tx.Value {
			continue
		}

		balances.Decrease(tx.Sender, tx.Value)
		balances.Increase(tx.Receiver, tx.Value)

		txs = append(candidates, tx)
	}

	return blockchain.NewBlock(prevHash, txs, difficulty, app.address.PublicKey())
}

func (app *application) processMinedBlocks() {
	var b *blockchain.Block
	for {
		app.updateMiningTarget()

		b = <-app.miner.MinedBlocks

		if err := app.ledger.AddBlock(*b); err != nil {
			app.logger.Error("Locally mined block is invalid", "error", err)
			continue
		}

		app.node.Broadcast(gossip.Message{
			Type: msgNewBlock,
			Data: b,
		})

		app.logger.Info("Mined and broadcasted a new block!")
	}
}
