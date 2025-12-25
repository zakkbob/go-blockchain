package gossip

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

var (
	ErrNoPeers = errors.New("node has no peers")
)

const maxPeers = 10

type Node struct {
	addr            string
	pubkey          ed25519.PublicKey
	logger          *slog.Logger
	receivedUpdates map[[32]byte]struct{}
	updateHandler   func(ReceivedUpdate) error
	requestHandler  func(ReceivedRequest) (response any, err error)
	listener        net.Listener
	peers           []*Peer
	mu              sync.RWMutex
}

func NewNode(addr string, pubkey ed25519.PublicKey, logger *slog.Logger, updateHandler func(ReceivedUpdate) error, requestHandler func(ReceivedRequest) (any, error)) *Node {
	return &Node{
		addr:            addr,
		pubkey:          pubkey,
		logger:          logger,
		receivedUpdates: map[[32]byte]struct{}{},
		updateHandler:   updateHandler,
		requestHandler:  requestHandler,
	}
}

func (n *Node) Connect(peers []string) error {
	return n.connectTo(peers)
}

func (n *Node) Listen() error {
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

		p, err := PeerFromConn(ctx, n.pubkey, c, n.handleUpdate, n.handleRequest)
		cancel()
		if err != nil {
			n.logger.Error("Failed to handle new connection", "error", err)
			continue
		}

		n.logger.Info("Accepted incoming connection", "RemoteAddr", c.RemoteAddr().String(), "Pubkey", fmt.Sprintf("%x", p.RemotePubkey))

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
				n.logger.Info("Disconnected from peer", "Pubkey", fmt.Sprintf("%x", p.RemotePubkey))
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

func (n *Node) Request(ctx context.Context, requestType string, data any) (ReceivedResponse, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.peers) == 0 {
		return ReceivedResponse{}, ErrNoPeers
	}

	res, err := n.peers[0].Request(ctx, requestType, data)
	if err != nil {
		return ReceivedResponse{}, err
	}

	return res, nil
}

func (n *Node) ListenerAddr() net.Addr {
	return n.listener.Addr()
}

func (n *Node) connectTo(addrs []string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var errs []error

	for _, addr := range addrs {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		peer, err := Dial(ctx, addr, n.pubkey, n.handleUpdate, n.handleRequest)
		cancel()
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
