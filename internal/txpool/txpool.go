package txpool

import "github.com/zakkbob/go-blockchain/internal/blockchain"

type Pool struct {
	txs []blockchain.Transaction
}

func (p *Pool) Size() int {
	return len(p.txs)
}

func (p *Pool) Add(tx blockchain.Transaction) {
	p.txs = append(p.txs, tx)
}

func (p *Pool) Get(n int) []blockchain.Transaction {
	n = min(n, len(p.txs))
	txs := p.txs[:n]
	p.txs = p.txs[n:]
	return txs
}
