package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

type config struct {
	debug bool
}

type application struct {
	config config
	logger *slog.Logger
	ledger *blockchain.Ledger
	node   *gossip.Node
}

func main() {
	port := flag.Int("port", 4000, "API server port")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ledger, err := blockchain.NewLedger(10)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	node := &gossip.Node{
		Addr:     ":3141",
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	app := application{
		config: config{
			debug: true,
		},
		logger: logger,
		ledger: ledger,
		node:   node,
	}

	logger.Info("starting server", "port", *port, "hash", ledger.Head().Hash())

	err = node.BootstrapAndListen([]string{}, app.handler)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
