package main

import (
	"encoding/json"
	"errors"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) handleUpdate(u gossip.ReceivedUpdate) error {
	var err error

	switch u.Type {
	case newBlockUpdate:
		err = app.handleNewBlock(u)
	case newTransactionUpdate:
		err = app.handleNewTransaction(u)
	default:
		app.logger.Error("Unknown update received", "update", u)
		return gossip.ErrUnknownUpdate
	}

	if err != nil {
		if errors.Is(err, gossip.ErrBadUpdate) || errors.Is(err, gossip.ErrUpdateRejected) {
			return err
		}
		app.logError("Failed to process update", err)
		return gossip.ErrUpdateRejected
	}

	return nil
}

func (app *application) handleNewTransaction(u gossip.ReceivedUpdate) error {
	var tx blockchain.Transaction

	err := json.Unmarshal(u.Data, &tx)
	if err != nil {
		app.logger.Info("Incoming transaction could not be parsed", "error", err)
		return gossip.ErrBadUpdate
	}

	if err = tx.Verify(); err != nil {
		app.logger.Info("Transaction rejected", "error", err)
		return gossip.ErrUpdateRejected
	}

	app.logger.Info("New transaction received", "transaction", tx.String())
	app.txpool.Add(tx)
	return nil
}

func (app *application) handleNewBlock(u gossip.ReceivedUpdate) error {
	var b blockchain.Block

	err := json.Unmarshal(u.Data, &b)
	if err != nil {
		app.logger.Info("Incoming block could not be parsed", "error", err)
		return gossip.ErrBadUpdate
	}

	err = app.ledger.AddBlock(b)
	if err != nil {
		app.logger.Info("Block rejected", "error", err)
		return gossip.ErrUpdateRejected
	}

	app.logger.Info("New block received")
	return nil
}
