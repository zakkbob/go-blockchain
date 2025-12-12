package gossip

import (
	"context"
	"encoding/json"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func receivedUpdateEquals(got ReceivedUpdate, expectedType string, expectedData any) bool {
	j, err := json.Marshal(expectedData)
	if err != nil {
		panic("can't marshal data")
	}

	return got.Type == expectedType && slices.Equal(got.Data, j)
}

func receivedRequestEquals(got ReceivedRequest, expectedType string, expectedData any) bool {
	j, err := json.Marshal(expectedData)
	if err != nil {
		panic("can't marshal data")
	}

	return got.Type == expectedType && slices.Equal(got.Data, j)
}

func receivedResponseEquals(got ReceivedResponse, expectedData any) bool {
	j, err := json.Marshal(expectedData)
	if err != nil {
		panic("can't marshal data")
	}

	return slices.Equal(got.Data, j)
}

func expectNoUpdate(t *testing.T) func(ReceivedUpdate) error {
	return func(u ReceivedUpdate) error {
		t.Errorf("Received unexpected update - %v", u)
		return nil
	}
}

func expectNoRequest(t *testing.T) func(ReceivedRequest) (any, error) {
	return func(r ReceivedRequest) (any, error) {
		t.Errorf("Received unexpected request - %v", r)
		return nil, nil
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
		return nil, nil
	}

}

type updateSpy struct { // bad name :/
	T          *testing.T
	LastUpdate ReceivedUpdate
}

func (s *updateSpy) HandleUpdate(u ReceivedUpdate) error {
	s.T.Log("Received Update - Type:", u.Type, "Data:", string(u.Data))
	s.LastUpdate = u

	return nil
}

type requestSpy struct { // bad name :/
	T           *testing.T
	LastRequest ReceivedRequest
	Response    any
}

func (s *requestSpy) ReceiveRequest(r ReceivedRequest) (any, error) {
	s.T.Logf("Received Request - %v", r)
	s.LastRequest = r
	return s.Response, nil
}

func TestPeerUpdate(t *testing.T) {
	conn1, conn2 := net.Pipe()

	spy := updateSpy{T: t}

	peer1, err := peerFromConn(conn1, spy.HandleUpdate, expectNoRequest(t))
	require.NoError(t, err)

	peer2, err := peerFromConn(conn2, expectNoUpdate(t), expectNoRequest(t))
	require.NoError(t, err)

	peer2.Update("type", "data")

	assert.Eventually(t, func() bool {
		return receivedUpdateEquals(spy.LastUpdate, "type", "data")
	}, time.Second, time.Millisecond, "Expected update was never received")

	peer1.Disconnect()
	peer2.Disconnect()
}

func TestPeerRequest(t *testing.T) {
	conn1, conn2 := net.Pipe()

	spy := requestSpy{
		T:        t,
		Response: "response",
	}

	peer1, err := peerFromConn(conn1, expectNoUpdate(t), spy.ReceiveRequest)
	require.NoError(t, err)

	peer2, err := peerFromConn(conn2, expectNoUpdate(t), expectNoRequest(t))
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Go(func() {
		res, err := peer2.Request(context.Background(), "type", "data")
		require.NoError(t, err)

		assert.Truef(t, receivedResponseEquals(res, "response"), "Received incorrect response - %v", res)

	})

	wg.Go(func() {
		assert.Eventually(t, func() bool {
			return receivedRequestEquals(spy.LastRequest, "type", "data")
		}, time.Second, time.Millisecond, "Expected request was never received")
	})

	wg.Wait()

	peer1.Disconnect()
	peer2.Disconnect()
}
