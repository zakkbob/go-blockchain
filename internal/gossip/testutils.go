package gossip

import (
	"encoding/json"
	"testing"
)

func NewReceivedUpdate(t *testing.T, updateType string,data any) ReceivedUpdate {
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal("Failed to create dummy received update")
		return ReceivedUpdate{}
	}

	return ReceivedUpdate{
		Type:       updateType,
		Data:       json.RawMessage(b),
	}
}
