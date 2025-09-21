package gossip

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/netip"
)

const port = 3141

type Message struct {
	RemoteAddr netip.AddrPort
	Type       string          `json:"message_type"`
	Data       json.RawMessage `json:"data"`
}

type Node struct {
	Addr     string
	ErrorLog *log.Logger
	listener *net.TCPListener
}

func (n *Node) Bootstrap(knownPeers []string) error {
	ip, err := netip.ParseAddr(n.Addr)
	if err != nil {
		return err
	}
	addr := net.TCPAddrFromAddrPort(netip.AddrPortFrom(ip, port))
	n.listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) Listen(handler func(string, json.RawMessage)) {
	for {
		c, err := n.listener.AcceptTCP()
		if err != nil {
			n.ErrorLog.Printf("Failed to accept incoming connection: %v", err)
			return
		}

		go n.handle(c, handler)
	}

}

func (n *Node) handle(conn *net.TCPConn, handler func(string, json.RawMessage)) {
	reader, writer := io.Pipe()
	go func() {
		_, err := conn.WriteTo(writer)
		writer.Close()
		if err != nil {
			n.ErrorLog.Printf("Error reading from TCPConn: %v", err)
		}
	}()

	d := json.NewDecoder(reader)

	for {
		m := Message{}

		err := d.Decode(&m)
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			n.ErrorLog.Printf("Failed to decode message from %s: %v", conn.RemoteAddr().String(), err)
			continue
		}

		handler(m.Type, m.Data)
	}
}
