package gossip_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/zakkbob/go-blockchain/internal/gossip"
)

func logHandler(messageType string, data json.RawMessage) {
	fmt.Print(messageType, "lol")
}

func TestBootstrap(t *testing.T) {
	n := gossip.Node{
		Addr:     "127.0.0.1",
		ErrorLog: log.Default(),
	}

	n.Bootstrap([]string{})
	go func() {
		n.Listen(logHandler)
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

	conn.Write([]byte(`{"message_type":steve,"data":""}`))

	time.Sleep(time.Second * 2)

	conn.Close()
}
