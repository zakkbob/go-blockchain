package gossip_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func TestBootstrap(t *testing.T) {
	recorder1 := updateRecorder{}
	recorder2 := updateRecorder{}

	n := gossip.NewNode(":0", slog.New(slog.DiscardHandler), recorder1.RecordUpdate, expectNoRequest(t))

	go func() {
		err := n.Listen()
		require.NoError(t, err)
	}()

	time.Sleep(time.Millisecond)
	t.Log(n.ListenerAddr())

	peer, err := gossip.Dial(n.ListenerAddr().String(), recorder2.RecordUpdate, expectNoRequest(t))
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
