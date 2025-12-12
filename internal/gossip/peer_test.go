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

type updateRecorder struct {
	updates []ReceivedUpdate
	mu      sync.Mutex
}

func (s *updateRecorder) NextUpdate() (ReceivedUpdate, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.updates) == 0 {
		return ReceivedUpdate{}, false
	}

	u := s.updates[0]
	s.updates = s.updates[1:]
	return u, true
}

func (s *updateRecorder) RecordUpdate(u ReceivedUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, u)
	return nil
}

type requestRecorder struct {
	Response any
	requests []ReceivedRequest
	mu       sync.Mutex
}

func (r *requestRecorder) NextRequest() (ReceivedRequest, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.requests) == 0 {
		return ReceivedRequest{}, false
	}

	u := r.requests[0]
	r.requests = r.requests[1:]
	return u, true
}

func (s *requestRecorder) RecordRequest(r ReceivedRequest) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, r)
	return s.Response, nil
}

func TestPeerUpdate(t *testing.T) {
	conn1, conn2 := net.Pipe()

	recorder := updateRecorder{}

	sender, err := peerFromConn(conn1, expectNoUpdate(t), expectNoRequest(t))
	require.NoError(t, err)

	receiver, err := peerFromConn(conn2, recorder.RecordUpdate, expectNoRequest(t))
	require.NoError(t, err)

	sender.Update("type", "data")
	assert.Eventually(t, func() bool {
		u, ok := recorder.NextUpdate()
		if !ok {
			return false
		}

		if !receivedUpdateEquals(u, "type", "data") {
			t.Errorf("Received unexpected update - %v", u)
			return false
		}

		return true
	}, time.Second, time.Millisecond, "Expected update was never received")

	sender.Disconnect()
	receiver.Disconnect()
}

func TestPeerRequest(t *testing.T) {
	conn1, conn2 := net.Pipe()

	recorder := requestRecorder{
		Response: "response",
	}

	sender, err := peerFromConn(conn2, expectNoUpdate(t), expectNoRequest(t))
	require.NoError(t, err)

	receiver, err := peerFromConn(conn1, expectNoUpdate(t), recorder.RecordRequest)
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Go(func() {
		res, err := sender.Request(context.Background(), "type", "data")
		require.NoError(t, err)

		assert.Truef(t, receivedResponseEquals(res, "response"), "Received incorrect response - %v", res)

	})

	wg.Go(func() {
		assert.Eventually(t, func() bool {
			r, ok := recorder.NextRequest()
			if !ok {
				return false
			}

			if !receivedRequestEquals(r, "type", "data") {
				t.Errorf("Received unexpected request - %v", r)
				return false
			}

			return true
		}, time.Second, time.Millisecond, "Expected request was never received")
	})

	wg.Wait()

	sender.Disconnect()
	receiver.Disconnect()
}
