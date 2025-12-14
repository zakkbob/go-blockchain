package main

import (
	"errors"

	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) handleRequest(r gossip.ReceivedRequest) (any, error) {
	var (
		res any
		err error
	)

	switch r.Type {
	case chainRequest:
		res, err = app.handleChainRequest(r)
	default:
		app.logger.Error("Unknown request received", "request", r)
		return nil, gossip.ErrUnknownRequest
	}

	if err != nil {
		if errors.Is(err, gossip.ErrBadRequest) {
			return nil, err
		}
		app.logError("Failed to process request", err)
		return nil, gossip.ErrBadRequest
	}

	return res, nil
}

func (app *application) handleChainRequest(r gossip.ReceivedRequest) (any, error) {
	return app.ledger.Chain(), nil
}
