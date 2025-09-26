package main

import (
	"encoding/json"
	"errors"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) handler(m gossip.ReceivedMessage) {
	switch m.Type {
	case msgNewBlock:
		app.newBlockHandler(m)
	case msgNewTransaction:
		app.newTransactionHandler(m)
	}

}

func (app *application) newTransactionHandler(m gossip.ReceivedMessage) {
	var tx blockchain.Transaction

	err := json.Unmarshal(m.Data, &tx)
	if err != nil {
		app.serverError(m, err)
		return
	}

	if err = tx.Verify(); err != nil {
		app.logger.Info("Transaction rejected", "remoteAddr", m.RemoteAddr, "error", err)
		return
	}

	app.logger.Info("New transaction received", "remoteAddr", m.RemoteAddr, "transaction", tx.String())
	app.txpool.Add(tx)
}

func (app *application) newBlockHandler(m gossip.ReceivedMessage) {
	var b blockchain.Block

	err := json.Unmarshal(m.Data, &b)
	if err != nil {
		app.serverError(m, err)
		return
	}

	err = app.ledger.AddBlock(b)
	var epbnf blockchain.ErrPrevBlockNotFound
	if errors.As(err, &epbnf) {
		app.logger.Info("Block rejected", "remoteAddr", m.RemoteAddr, "error", err)
		return
	} else if err != nil {
		app.serverError(m, err)
		return
	}

	app.logger.Info("New block received", "remoteAddr", m.RemoteAddr)
}
