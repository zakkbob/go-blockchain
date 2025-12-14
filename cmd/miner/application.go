package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func (app *application) run(peers []string, difficulty int) error {
	app.logger.Info("Starting server", "port", port)

	if len(peers) > 0 {
		err := app.node.Connect(peers)
		if err != nil {
			return err
		}

		var blocks []blockchain.Block

		res, err := app.node.Request(context.TODO(), chainRequest, nil)
		if err != nil {
			return fmt.Errorf("failed to sync chain: %w", err)
		}

		err = json.Unmarshal(res.Data, &blocks)
		if err != nil {
			return err
		}

		for i := len(blocks) - 1; i >= 0; i-- {
			app.ledger.AddBlock(blocks[i])
		}

		app.logger.Info("Synced blockchain with peer", "chain", res.String())
	} else {
		b := blockchain.NewGenesisBlock(difficulty)
		b.Mine()
		app.ledger.AddBlock(b)
		app.logger.Info("Created new genesis block", "block", b)
	}

	go app.processMinedBlocks()

	return app.node.Listen()
}
