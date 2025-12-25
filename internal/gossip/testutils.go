package gossip

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zakkbob/go-blockchain/internal/blockchain"
)

func CreatePeerPipe(t *testing.T, peerOneUpdateHandler func(ReceivedUpdate) error, peerOneRequestHandler func(ReceivedRequest) (any, error), peerTwoUpdateHandler func(ReceivedUpdate) error, peerTwoRequestHandler func(ReceivedRequest) (any, error)) (*Peer, *Peer) {
	conn1, conn2 := net.Pipe()
	addr1 := blockchain.MustGenerateTestAddress(t)
	addr2 := blockchain.MustGenerateTestAddress(t)

	var peerOne, peerTwo *Peer

	wg := sync.WaitGroup{}

	wg.Go(func() {
		var err error
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		peerOne, err = PeerFromConn(ctx, addr1.PublicKey(), conn1, peerOneUpdateHandler, peerOneRequestHandler)
		cancel()
		require.NoError(t, err)
	})

	wg.Go(func() {
		var err error
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		peerTwo, err = PeerFromConn(ctx, addr2.PublicKey(), conn2, peerTwoUpdateHandler, peerTwoRequestHandler)
		cancel()
		require.NoError(t, err)

	})

	wg.Wait()

	return peerOne, peerTwo
}

func NewReceivedUpdate(t *testing.T, updateType string, data any) ReceivedUpdate {
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal("Failed to create dummy received update")
		return ReceivedUpdate{}
	}

	return ReceivedUpdate{
		Type: updateType,
		Data: json.RawMessage(b),
	}
}
