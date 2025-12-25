package gossip_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zakkbob/go-blockchain/internal/blockchain"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func TestBootstrap(t *testing.T) {
	recorder1 := updateRecorder{}
	recorder2 := updateRecorder{}

	addr1 := blockchain.MustGenerateTestAddress(t)
	addr2 := blockchain.MustGenerateTestAddress(t)

	n := gossip.NewNode(":0", addr1.PublicKey(), slog.New(slog.DiscardHandler), recorder1.RecordUpdate, expectNoRequest(t))

	go func() {
		err := n.Listen()
		require.NoError(t, err)
	}()

	time.Sleep(time.Millisecond)
	t.Log(n.ListenerAddr())

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	peer, err := gossip.Dial(ctx, n.ListenerAddr().String(), addr2.PublicKey(), recorder2.RecordUpdate, expectNoRequest(t))
	cancel()
	require.NoError(t, err)

	err = peer.Update("type", "data")
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		u, ok := recorder1.NextUpdate()
		if !ok {
			return false
		}

		if !receivedUpdateEquals(u, "type", "data") {
			t.Errorf("Received unexpected update - %v", u)
			return false
		}

		return true
	}, time.Second, time.Millisecond, "Never received expected update")

	assert.Eventually(t, func() bool {
		u, ok := recorder2.NextUpdate()
		if !ok {
			return false
		}

		if !receivedUpdateEquals(u, "type", "data") {
			t.Errorf("Received unexpected update - %v", u)
			return false
		}

		return true
	}, time.Second, time.Millisecond, "Never received expected update (through gossiping)")

	peer.Disconnect()
}
