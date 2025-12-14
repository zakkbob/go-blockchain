package gossip

import (
	"errors"
	"log/slog"
	"net"
	"sync"
)

const maxPeers = 10

type Node struct {
	addr            string
	logger          *slog.Logger
	receivedUpdates map[[32]byte]struct{}
	updateHandler   func(ReceivedUpdate) error
	requestHandler  func(ReceivedRequest) (response any, err error)
	listener        net.Listener
	peers           []*Peer
	mu              sync.RWMutex
}

func NewNode(addr string, logger *slog.Logger) *Node {
	return &Node{
		addr:            addr,
		logger:          logger,
		receivedUpdates: map[[32]byte]struct{}{},
	}
}

func (n *Node) BootstrapAndListen(knownPeers []string, updateHandler func(ReceivedUpdate) error, requestHandler func(ReceivedRequest) (any, error)) error {
	n.updateHandler = updateHandler
	n.requestHandler = requestHandler

	n.connectTo(knownPeers)

	var err error

	n.listener, err = net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}

	for {
		c, err := n.listener.Accept()
		if err != nil {
			n.logger.Error("Failed to accept incoming connection", "error", err)
			continue
		}

		n.mu.RLock()
		nPeers := len(n.peers)
		n.mu.RUnlock()

		if nPeers >= maxPeers {
			n.logger.Info("Rejected incoming connection, already at max peers", "address", n.addr)
			c.Close()
			continue
		}

		n.logger.Info("Accepted incoming connection", "RemoteAddr", c.RemoteAddr().String())

		p, err := PeerFromConn(c, n.handleUpdate, n.handleRequest)
		if err != nil {
			n.logger.Error("Failed to handle new successful connection", "error", err)
		}

		n.mu.Lock()
		n.peers = append(n.peers, p)
		n.mu.Unlock()
	}
}

func (n *Node) BroadcastUpdate(updateType string, data any) error {
	var errs []error
	var disconnectedPeers = false

	n.mu.RLock()
	for _, p := range n.peers {
		if p.Status == Disconnected {
			disconnectedPeers = true
			continue
		}

		err := p.Update(updateType, data)
		if err != nil {
			if errors.Is(err, ErrPeerDisconnected) {
				disconnectedPeers = true
				continue
			} else {
				errs = append(errs, err)
			}
		}
	}
	n.mu.RUnlock()

	if disconnectedPeers {
		n.mu.Lock()
		i := 0
		for _, p := range n.peers {
			if p.Status == Connected {
				n.peers[i] = p
				i++
			} else {
				n.logger.Debug("Removed disconnected peer from list", "RemoteAddr", p.RemoteAddr)
			}
		}
		n.peers = n.peers[:i]
		n.mu.Unlock()
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (n *Node) ListenerAddr() net.Addr {
	return n.listener.Addr()
}

func (n *Node) connectTo(addrs []string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var errs []error

	for _, addr := range addrs {
		peer, err := Dial(addr, n.handleUpdate, n.handleRequest)
		if err != nil {
			n.logger.Error("Failed to connect to peer", "peer", addr, "error", err)
			errs = append(errs, err)
			continue
		}
		n.peers = append(n.peers, peer)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (n *Node) handleUpdate(u ReceivedUpdate) error {
	h := u.Hash()
	if _, ok := n.receivedUpdates[h]; ok {
		n.logger.Debug("Ignored already received update", "update", u)
		return nil
	}

	n.receivedUpdates[h] = struct{}{}

	err := n.updateHandler(u)
	if err != nil && !errors.Is(err, ErrUpdateRejected) {
		return err
	}

	n.BroadcastUpdate(u.Type, u.Data)
	return nil
}

func (n *Node) handleRequest(r ReceivedRequest) (any, error) {
	return n.requestHandler(r)
}
