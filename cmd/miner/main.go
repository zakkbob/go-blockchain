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

type peersFlag []string

func (p *peersFlag) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p *peersFlag) Set(value string) error {
	*p = append(*p, value)
	return nil
}

var debugFlag bool
var port int
var difficulty int
var peers peersFlag

func main() {
	flag.BoolVar(&debugFlag, "debug", false, "Show debug info")
	flag.IntVar(&port, "port", 4000, "API server port")
	flag.IntVar(&difficulty, "difficulty", 10, "Mining difficulty")
	flag.Var(&peers, "peer", "Peers (can be used multiple times)")

	flag.Parse()

	level := slog.LevelInfo
	if debugFlag {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	ledger, err := blockchain.NewLedger(difficulty)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	node := gossip.NewNode(fmt.Sprintf(":%d", port), logger)

	address, err := blockchain.GenerateAddress(rand.Reader)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	miner := miner.NewMiner(address.PublicKey())

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

	logger.Info("starting server", "port", port, "hash", ledger.Head().Hash())

	err = node.BootstrapAndListen(peers, app.handleUpdate, app.handleRequest)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
