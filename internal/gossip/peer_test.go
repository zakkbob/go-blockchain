package gossip

import (
	"encoding/json"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func receivedUpdateEquals(got ReceivedUpdate, expectedType string, expectedData any) bool {
	j, err := json.Marshal(expectedData)
	if err != nil {
		panic("can't marshal data")
	}

	return got.Type == expectedType && slices.Equal(got.Data, j)
}

func logReceivedUpdate(t *testing.T) func(ReceivedUpdate) error {
	return func(u ReceivedUpdate) error {
		t.Log("Received Update - Type:", u.Type, "Data:", string(u.Data))
		return nil
	}
}

func logReceivedRequest(t *testing.T) func(ReceivedRequest) (any, error) {
	return func(r ReceivedRequest) (any, error) {
		t.Log("Received Request - Type:", r.Type, "Data:", string(r.Data))
		return 1, nil
	}

}

type updateSpy struct { // bad name :/
	t          *testing.T
	LastUpdate ReceivedUpdate
}

func (s *updateSpy) HandleUpdate(u ReceivedUpdate) error {
	s.t.Log("Received Update - Type:", u.Type, "Data:", string(u.Data))
	s.LastUpdate = u

	return nil
}

type requestSpy struct { // bad name :/
	LastRequest ReceivedRequest
	Response    any
}

func (s *requestSpy) ReceiveRequest(r ReceivedRequest) (any, error) {
	s.LastRequest = r
	return s.Response, nil
}

func TestPeerUpdate(t *testing.T) {
	conn1, conn2 := net.Pipe()

	uSpy1 := updateSpy{t: t}

	peer1, err := peerFromConn(conn1, uSpy1.HandleUpdate, logReceivedRequest(t))
	if err != nil {
		t.Fatal(err.Error())
	}

	peer2, err := peerFromConn(conn2, logReceivedUpdate(t), logReceivedRequest(t))
	if err != nil {
		t.Fatal(err.Error())
	}

	peer2.Update("type", "data")

	assert.Eventually(t, func() bool {
		return receivedUpdateEquals(uSpy1.LastUpdate, "type", "data")
	}, time.Second, time.Millisecond, "Expected update was never received")

	peer1.Disconnect()
	peer2.Disconnect()
}

// test requests, maybe fix up the update test a bit better, cleaner
