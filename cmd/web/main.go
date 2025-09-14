package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

type config struct {
	debug bool
}

type application struct {
	config config
	logger *slog.Logger
	ledger *blockchain.Ledger
}

func main() {
	port := flag.Int("port", 4000, "API server port")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ledger, err := blockchain.NewLedger(10)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := application{
		config: config{
			debug: true,
		},
		logger: logger,
		ledger: ledger,
	}

	logger.Info("starting server", "port", *port, "hash", ledger.Head().Hash())

	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), app.routes())
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
