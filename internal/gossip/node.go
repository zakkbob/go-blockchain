package gossip

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
)

type ReceivedMessage struct {
	Type       string
	RemoteAddr string
	Data       json.RawMessage
}

func (m *ReceivedMessage) Hash() [32]byte {
	b := make([]byte, 0, len(m.Type)+len(m.RemoteAddr)+len(m.Data))
	b = append(b, []byte(m.Type)...)
	b = append(b, []byte(m.RemoteAddr)...)
	b = append(b, []byte(m.Data)...)
	return sha256.Sum256(b)
}

type Message struct {
	Type string `json:"message_type"`
	Data any    `json:"data"`
}

type Node struct {
	Addr     string
	ErrorLog *log.Logger
	handler  func(ReceivedMessage)
	listener net.Listener
	conns    []net.Conn
}

func (n *Node) ListenerAddr() net.Addr {
	return n.listener.Addr()
}

func (n *Node) connectTo(knownPeers []string) error {
	var errs []error

	for _, peer := range knownPeers {
		conn, err := net.Dial("tcp", peer)
		if err != nil {
			n.ErrorLog.Printf("Failed to connect to peer %s: %v", peer, err)
			errs = append(errs, err)
			continue
		}

		go n.handle(conn)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (n *Node) Broadcast(m Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	for _, c := range n.conns {
		_, err = c.Write(b)
	}
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) BootstrapAndListen(knownPeers []string, handler func(ReceivedMessage)) error {
	n.handler = handler

	n.connectTo(knownPeers)

	var err error

	n.listener, err = net.Listen("tcp", n.Addr)
	if err != nil {
		return err
	}

	for {
		c, err := n.listener.Accept()
		if err != nil {
			n.ErrorLog.Printf("Failed to accept incoming connection: %v", err)
			continue
		}

		go n.handle(c)
	}
}

func (n *Node) handle(c net.Conn) {
	n.conns = append(n.conns, c)

	d := json.NewDecoder(c)

	for {
		m := struct {
			Type string          `json:"message_type"`
			Data json.RawMessage `json:"data"`
		}{}

		err := d.Decode(&m)
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			n.ErrorLog.Printf("Failed to decode message from %s: %v", c.RemoteAddr().String(), err)
			continue
		}

		r := ReceivedMessage{
			Type:       m.Type,
			Data:       m.Data,
			RemoteAddr: c.RemoteAddr().String(),
		}

		n.handler(r)
	}
}
