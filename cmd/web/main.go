package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
	"github.com/zakkbob/go-blockchain/internal/miner"
	"github.com/zakkbob/go-blockchain/internal/txpool"
)

type config struct {
	debug bool
}

type application struct {
	config  config
	address blockchain.Address
	miner   *miner.Miner
	logger  *slog.Logger
	ledger  *blockchain.Ledger
	node    *gossip.Node
	txpool  txpool.Pool
}

func main() {
	port := flag.Int("port", 4000, "API server port")
	difficulty := flag.Int("difficulty", 10, "Mining difficulty")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ledger, err := blockchain.NewLedger(*difficulty)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	node := &gossip.Node{
		Addr:     fmt.Sprintf(":%d", *port),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	address, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	miner := miner.NewMiner(address.PublicKey(), 4)

	app := application{
		config: config{
			debug: true,
		},
		address: address,
		logger:  logger,
		ledger:  ledger,
		miner:   miner,
		node:    node,
		txpool:  txpool.Pool{},
	}

	go app.processMinedBlocks()

	logger.Info("starting server", "port", *port, "hash", ledger.Head().Hash())

	err = node.BootstrapAndListen([]string{"[::]:4100"}, app.handler)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
