package main

import (
	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) updateMiningTarget() {
	b := app.ledger.ConstructNextBlock(map[[32]byte]blockchain.Transaction{}, app.address.PublicKey())
	app.miner.SetTargetBlock(b)
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
			Type: "newBlock",
			Data: b,
		})

		app.logger.Info("Mined and broadcasted a new block!")
	}
}
