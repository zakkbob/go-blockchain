package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodPost, "/v1/block", app.newBlockHandler)

	return router
}

func (app *application) newBlockHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("New block received", "host", r.Host, "remoteAddr", r.RemoteAddr)

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var b blockchain.Block

	err = json.Unmarshal(body, &b)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.ledger.AddBlock(b)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
