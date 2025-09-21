package gossip_test

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/zakkbob/go-blockchain/internal/gossip"
)

type testHandler struct {
	Type string
	Data json.RawMessage
}

func (t *testHandler) handle(messageType string, data json.RawMessage) {
	t.Type = messageType
	t.Data = data
}

func TestBootstrap(t *testing.T) {
	handler := testHandler{}

	n := gossip.Node{
		Addr:     "127.0.0.1",
		ErrorLog: slog.NewLogLogger(slog.DiscardHandler, slog.LevelDebug),
	}

	n.Bootstrap([]string{})
	go func() {
		n.Listen(handler.handle)
	}()

	lAddr, err := netip.ParseAddrPort("127.0.0.1:1234")
	if err != nil {
		t.Fatal(err)
	}

	rAddr, err := netip.ParseAddrPort("127.0.0.1:3141")
	if err != nil {
		t.Fatal(err)
	}

	lTCPAddr := net.TCPAddrFromAddrPort(lAddr)
	rTCPAddr := net.TCPAddrFromAddrPort(rAddr)

	conn, err := net.DialTCP("tcp4", lTCPAddr, rTCPAddr)
	if err != nil {
		t.Fatal(err)
	}

	conn.Write([]byte(`{"message_type":"steve","data":""}`))

	time.Sleep(time.Second * 1)

	if handler.Type != "steve" {
		t.Errorf("I need steve")
	}

	conn.Close()
}
