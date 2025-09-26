package gossip

import (
	"encoding/json"
	"testing"
)

func CreateReceivedMessage(t *testing.T, messageType string, remoteAddr string, data any) ReceivedMessage {
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal("Failed to create received message")
		return ReceivedMessage{}
	}

	return ReceivedMessage{
		Type:       messageType,
		RemoteAddr: remoteAddr,
		Data:       json.RawMessage(b),
	}
}
