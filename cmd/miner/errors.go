package main

import (
	"runtime/debug"
)

func (app *application) logError(err error) {
	var (
		trace = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "trace", trace)
}

func (app *application) serverError(err error) {
	app.logError(err)
}
