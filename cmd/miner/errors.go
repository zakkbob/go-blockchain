package main

import (
	"runtime/debug"
)

func (app *application) logError(msg string, err error) {
	var (
		trace = string(debug.Stack())
	)

	app.logger.Error(msg, "error", err, "trace", trace)
}
