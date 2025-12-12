package main

import (
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) handleRequest(r gossip.ReceivedRequest) (any, error) {
	app.logger.Error("Unknown request received", "request", r)
	return nil, gossip.ErrUnknownRequest
}
