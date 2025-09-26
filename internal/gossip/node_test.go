package gossip_test

import (
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/zakkbob/go-blockchain/internal/gossip"
)

type testHandler struct {
	message gossip.ReceivedMessage
}

func (t *testHandler) handle(message gossip.ReceivedMessage) {
	t.message = message
}

func TestBootstrap(t *testing.T) {
	handler := testHandler{}

	n := gossip.Node{
		Addr:     ":0",
		ErrorLog: slog.NewLogLogger(slog.DiscardHandler, slog.LevelDebug),
	}

	go func() {
		err := n.BootstrapAndListen([]string{}, handler.handle)
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Millisecond)
	t.Log(n.ListenerAddr())

	conn, err := net.Dial("tcp", n.ListenerAddr().String())
	if err != nil {
		t.Fatal(err)
	}

	conn.Write([]byte(`{"message_type":"steve","data":""}`))

	time.Sleep(time.Millisecond)

	if handler.message.Type != "steve" {
		t.Errorf("I need steve")
	}

	conn.Close()
}
