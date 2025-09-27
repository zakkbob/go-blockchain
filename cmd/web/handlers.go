package main

import (
	"encoding/json"
	"errors"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) handler(m gossip.ReceivedMessage) {
	_, ok := app.receivedMessages[m.Hash()]
	if ok {
		return
	}

	rebroadcast := false

	switch m.Type {
	case msgNewBlock:
		rebroadcast = app.newBlockHandler(m)
	case msgNewTransaction:
		rebroadcast = app.newTransactionHandler(m)
	default:
		app.logger.Error("Unknown message received", "message", m)
	}

	if rebroadcast {
		app.node.Broadcast(gossip.Message{
			Type: m.Type,
			Data: m.Data,
		})
	}
}

func (app *application) newTransactionHandler(m gossip.ReceivedMessage) bool {
	var tx blockchain.Transaction

	err := json.Unmarshal(m.Data, &tx)
	if err != nil {
		app.serverError(m, err)
		return false
	}

	if err = tx.Verify(); err != nil {
		app.logger.Info("Transaction rejected", "remoteAddr", m.RemoteAddr, "error", err)
		return false
	}

	app.logger.Info("New transaction received", "remoteAddr", m.RemoteAddr, "transaction", tx.String())
	app.txpool.Add(tx)
	return true
}

func (app *application) newBlockHandler(m gossip.ReceivedMessage) bool {
	var b blockchain.Block

	err := json.Unmarshal(m.Data, &b)
	if err != nil {
		app.serverError(m, err)
		return false
	}

	err = app.ledger.AddBlock(b)
	var epbnf blockchain.ErrPrevBlockNotFound
	if errors.As(err, &epbnf) {
		app.logger.Info("Block rejected", "remoteAddr", m.RemoteAddr, "error", err)
		return false
	} else if err != nil {
		app.serverError(m, err)
		return false
	}

	app.logger.Info("New block received", "remoteAddr", m.RemoteAddr)
	return true
}
