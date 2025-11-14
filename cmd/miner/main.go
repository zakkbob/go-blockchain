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
	config           config
	address          blockchain.Address
	miner            *miner.Miner
	logger           *slog.Logger
	ledger           *blockchain.Ledger
	node             *gossip.Node
	txpool           txpool.Pool
	receivedMessages map[[32]byte]struct{}
}

type peersFlag []string

func (p *peersFlag) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p *peersFlag) Set(value string) error {
	*p = append(*p, value)
	return nil
}

var port int
var difficulty int
var peers peersFlag

func main() {
	flag.IntVar(&port, "port", 4000, "API server port")
	flag.IntVar(&difficulty, "difficulty", 10, "Mining difficulty")
	flag.Var(&peers, "peer", "Peers (can be used multiple times)")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ledger, err := blockchain.NewLedger(difficulty)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	node := &gossip.Node{
		Addr:   fmt.Sprintf(":%d", port),
		Logger: logger,
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
		address:          address,
		logger:           logger,
		ledger:           ledger,
		miner:            miner,
		node:             node,
		txpool:           txpool.Pool{},
		receivedMessages: map[[32]byte]struct{}{},
	}

	go app.processMinedBlocks()

	logger.Info("starting server", "port", port, "hash", ledger.Head().Hash())

	err = node.BootstrapAndListen(peers, app.handler)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
