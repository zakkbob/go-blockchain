package gossip

import (
	"net"
	"slices"
	"testing"
	"time"
)

func assertReceivedUpdateEqual(t *testing.T, got ReceivedUpdate, expected ReceivedUpdate) {
	if got.Type != expected.Type || !slices.Equal(got.Data, expected.Data) {
		t.Fatalf("Expected update %v, but got %v", expected, got)
	}

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
	time.Sleep(time.Millisecond)
	assertReceivedUpdateEqual(t, uSpy1.LastUpdate, ReceivedUpdate{
		Type: "type",
		Data: []byte("\"data\""),
	})

	peer1.Disconnect()
	peer2.Disconnect()
}
