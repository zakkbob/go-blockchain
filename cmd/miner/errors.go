package main

import (
	"runtime/debug"

	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func (app *application) logError(m gossip.ReceivedMessage, err error) {
	var (
		trace = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "trace", trace)
}

func (app *application) serverError(m gossip.ReceivedMessage, err error) {
	app.logError(m, err)
}
