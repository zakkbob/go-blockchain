package gossip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"sync"
	"time"
)

var (
	ErrNoPeers = errors.New("node has no peers")
)

const maxPeers = 10

type Node struct {
	listener        net.Listener
	addr            string // local address
	logger          *slog.Logger
	receivedUpdates map[[32]byte]struct{} // hashes of received updates
	updateHandler   func(ReceivedUpdate) error
	requestHandler  func(ReceivedRequest) (response any, err error)
	peers           []*Peer // connected peers
	peerAddrs       map[string]struct{}
	peersMu         sync.RWMutex // mutex for peers array and map
	knownPeers      []string     // addresses of potential future peers

}

func NewNode(addr string, logger *slog.Logger, updateHandler func(ReceivedUpdate) error, requestHandler func(ReceivedRequest) (any, error)) *Node {
	return &Node{
		addr:            addr,
		logger:          logger,
		receivedUpdates: map[[32]byte]struct{}{},
		updateHandler:   updateHandler,
		requestHandler:  requestHandler,
		peerAddrs:       map[string]struct{}{},
	}
}

func (n *Node) Listen() error {
	var err error

	l, err := net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}

	n.listener = l

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			n.logger.Error("Failed to accept incoming connection", "error", err)
			c.Close()
			continue
		}

		go func() {
			n.peersMu.RLock()
			nPeers := len(n.peers)
			n.peersMu.RUnlock()

			if nPeers >= maxPeers {
				n.logger.Info("Rejected incoming connection, already at max peers", "address", n.addr)
				c.Close()
				return
			}

			n.logger.Info("Accepted incoming connection", "RemoteAddr", c.RemoteAddr().String())

			p, err := PeerFromConn(c, n.handleUpdate, n.handleRequest)
			if err != nil {
				n.logger.Error("Failed to handle new successful connection", "error", err)
				c.Close()
				return
			}

			n.handleNewPeer(p)
		}()
	}
}

func (n *Node) BroadcastUpdate(updateType string, data any) error {
	var errs []error
	var disconnectedPeers = false

	n.peersMu.RLock()
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
	n.peersMu.RUnlock()

	if disconnectedPeers {
		n.peersMu.Lock()
		i := 0
		for _, p := range n.peers {
			if p.Status == Connected {
				n.peers[i] = p
				i++
			} else {
				delete(n.peerAddrs, p.RemoteAddr)
				n.logger.Debug("Removed disconnected peer from list", "RemoteAddr", p.RemoteAddr)
			}
		}
		n.peers = n.peers[:i]
		n.peersMu.Unlock()
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (n *Node) Request(ctx context.Context, requestType string, data any) (ReceivedResponse, error) {
	n.peersMu.RLock()
	defer n.peersMu.RUnlock()

	if len(n.peers) == 0 {
		return ReceivedResponse{}, ErrNoPeers
	}

	res, err := n.peers[0].Request(ctx, requestType, data)
	if err != nil {
		return ReceivedResponse{}, err
	}

	n.logger.Debug("Received response to request", "requestType", requestType, "response", res)

	return res, nil
}

func (n *Node) ListenerAddr() net.Addr {
	return n.listener.Addr()
}

func (n *Node) Connect(addr string) error {
	if len(n.peers) >= maxPeers {
		if !slices.Contains(n.knownPeers, addr) {
			n.knownPeers = append(n.knownPeers, addr)
		}
		return nil
	}

	if addr == n.addr {
		return nil
	}

	p, err := Dial(addr, n.handleUpdate, n.handleRequest)
	if err != nil {
		return err
	}

	return n.handleNewPeer(p)
}

func (n *Node) handleNewPeer(p *Peer) error {
	n.peersMu.Lock()
	n.logger.Debug("Peers", "peers", n.peerAddrs)
	if _, ok := n.peerAddrs[p.RemoteAddr]; ok {
		n.peersMu.Unlock()
		n.logger.Debug("Prevented duplicate connection", "RemoteAddr", p.RemoteAddr)
		return nil
	}

	n.peers = append(n.peers, p)
	n.peerAddrs[p.RemoteAddr] = struct{}{}
	n.peersMu.Unlock()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	res, err := p.Request(ctx, peersRequest, nil)
	if err != nil {
		return fmt.Errorf("failed to request peer list from peer: %w", err)
	}

	var peers []string

	err = json.Unmarshal(res.Data, &peers)
	if err != nil {
		return fmt.Errorf("failed to decode peer list: %w", err)
	}

	n.knownPeers = append(n.knownPeers, peers...)
	n.attemptKnownPeers()

	n.logger.Info("Connected to new peer", "RemoteAddr", p.RemoteAddr)

	return nil
}

func (n *Node) attemptKnownPeers() error {
	for i, p := range n.knownPeers {
		n.peersMu.RLock()
		if len(n.peers) >= maxPeers {
			n.peersMu.RUnlock()
			n.knownPeers = n.knownPeers[i:]
			return nil
		}
		n.peersMu.RUnlock()

		err := n.Connect(p)
		if err != nil {
			n.knownPeers = n.knownPeers[i:]
			return err
		}
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
	n.logger.Info("Received request", "Type", r.Type)

	switch r.Type {
	case peersRequest:
		var peers []string
		for _, p := range n.peers {
			peers = append(peers, p.RemoteAddr)
		}
		return peers, nil
	}
	return n.requestHandler(r)
}
