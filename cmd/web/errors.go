package main

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
)

type envelope = map[string]any

func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, code int, msg string) {
	js := map[string]string{
		"message": msg,
	}

	body, err := json.Marshal(envelope{"error": js})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(code)
	w.Write(body)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request) {
	msg := "bad request"
	app.errorResponse(w, r, http.StatusBadRequest, msg)
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "internal server error"
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}
