package gossip

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	ErrUnknownUpdate      = errors.New("unkown update type")
	ErrBadUpdate          = errors.New("update data cannot be parsed")
	ErrUpdateRejected     = errors.New("update was rejected")
	ErrUnknownRequest     = errors.New("unkown request type")
	ErrBadRequest         = errors.New("request data cannot be parsed")
	ErrPeerDisconnected   = errors.New("peer has been disconnected")
	ErrUnexpectedResponse = errors.New("received response does not match a request")
)

type messageType int

const (
	update messageType = iota
	request
	response
)

type message struct {
	MessageType messageType `json:"message_type"`
	Message     any         `json:"message"`
}

type Update struct {
	Type string `json:"update_type"`
	Data any    `json:"data"`
}

type ReceivedUpdate struct {
	Type string          `json:"update_type"`
	Data json.RawMessage `json:"data"`
}

func (r ReceivedUpdate) Hash() [32]byte {
	b := []byte(r.Type)
	b = append(b, r.Data...)
	return sha256.Sum256(b)
}

func (r ReceivedUpdate) String() string {
	return fmt.Sprintf("{Type: '%s', Data: '%s'}", r.Type, string(r.Data))
}

type Request struct {
	ID   int    `json:"id"`
	Type string `json:"request_type"`
	Data any    `json:"data"`
}

type ReceivedRequest struct {
	ID   int             `json:"id"`
	Type string          `json:"request_type"`
	Data json.RawMessage `json:"data"`
}

func (r ReceivedRequest) String() string {
	return fmt.Sprintf("{ID: %d, Type: '%s', Data: '%s'}", r.ID, r.Type, string(r.Data))
}

type Response struct {
	RequestID int `json:"request_id"`
	Data      any `json:"data"`
}

type ReceivedResponse struct {
	RequestID int             `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

func (r ReceivedResponse) String() string {
	return fmt.Sprintf("{RequestID: %d, Data: '%s'}", r.RequestID, string(r.Data))
}

type status bool

const (
	Connected    = true
	Disconnected = false
)

type Peer struct {
	Status     status
	RemoteAddr string

	conn   net.Conn
	lastID atomic.Int64

	updateHandler  func(ReceivedUpdate) error
	requestHandler func(ReceivedRequest) (response any, err error)

	responseMap map[int]chan ReceivedResponse
	mu          sync.RWMutex

	closeErr error // isnt handled properlu yet
}

func Dial(address string, updateHandler func(ReceivedUpdate) error, requestHandler func(ReceivedRequest) (any, error)) (*Peer, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return PeerFromConn(conn, updateHandler, requestHandler)
}

func PeerFromConn(conn net.Conn, updateHandler func(ReceivedUpdate) error, requestHandler func(ReceivedRequest) (any, error)) (*Peer, error) {
	p := &Peer{
		Status:         Connected,
		RemoteAddr:     conn.RemoteAddr().String(),
		conn:           conn,
		updateHandler:  updateHandler,
		requestHandler: requestHandler,
		responseMap:    map[int]chan ReceivedResponse{},
	}

	go p.handle()

	return p, nil
}

func (p *Peer) Update(updateType string, data any) error {
	if p.closeErr != nil {
		return p.closeErr
	}

	return p.send(message{
		MessageType: update,
		Message: Update{
			Type: updateType,
			Data: data,
		},
	})
}

func (p *Peer) Request(ctx context.Context, requestType string, data any) (ReceivedResponse, error) {
	if p.closeErr != nil {
		return ReceivedResponse{}, p.closeErr
	}

	id := p.nextID()

	m := message{
		MessageType: request,
		Message: Request{
			ID:   id,
			Type: requestType,
			Data: data,
		},
	}

	resChan := p.registerRequestID(id)

	err := p.send(m)
	if err != nil {
		p.unregisterRequestID(id)
		return ReceivedResponse{}, err
	}

	select {
	case res, ok := <-resChan:
		if !ok {
			if p.closeErr != nil {
				return ReceivedResponse{}, p.closeErr
			}
			panic("response channel closed unexpectedly")
		}
		return res, nil
	case <-ctx.Done():
		p.unregisterRequestID(id)
		return ReceivedResponse{}, ctx.Err()
	}
}

func (p *Peer) Disconnect() error {
	if p.closeErr != nil {
		return p.closeErr
	}

	p.fatalError(ErrPeerDisconnected)

	return nil
}

// not proper yet
func (p *Peer) fatalError(err error) {
	p.conn.Close()
	p.closeErr = fmt.Errorf("peer already disconnected due to fatal error: %w", err)
	p.clearResponseMap()
	p.Status = Disconnected
}

func (p *Peer) clearResponseMap() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, resChan := range p.responseMap {
		close(resChan)
		delete(p.responseMap, id)
	}
}

func (p *Peer) registerRequestID(id int) chan ReceivedResponse {
	p.mu.Lock()
	defer p.mu.Unlock()
	resChan := make(chan ReceivedResponse)
	p.responseMap[id] = resChan
	return resChan
}

func (p *Peer) unregisterRequestID(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.responseMap[id])
	delete(p.responseMap, id)
}

func (p *Peer) nextID() int {
	return int(p.lastID.Add(1))
}

func (p *Peer) send(m message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	_, err = p.conn.Write(b)
	if err != nil {
		if errors.Is(err, syscall.EPIPE) { // Connection was closed from remote side
			p.Disconnect()
			return ErrPeerDisconnected
		}
		return err
	}

	return nil
}

func (p *Peer) handle() {
	var (
		u   ReceivedUpdate
		r   ReceivedRequest
		res ReceivedResponse
	)
	raw := &bytes.Buffer{}
	d := json.NewDecoder(io.TeeReader(p.conn, raw))

	for {
		m := struct {
			MessageType messageType     `json:"message_type"`
			Message     json.RawMessage `json:"message"`
		}{}

		err := d.Decode(&m)
		if errors.Is(err, io.EOF) { // Connection closed from remote side
			p.Disconnect()
			return
		} else if err != nil {
			p.fatalError(err)
			return
		} else if p.closeErr != nil {
			return
		}

		switch m.MessageType {
		case update:
			if err := json.Unmarshal(m.Message, &u); err != nil {
				p.fatalError(err)
				return
			}
			err = p.handleReceivedUpdate(u)
			if err != nil {
				p.fatalError(fmt.Errorf("failed to handle received update: %w", err))
				return
			}
		case request:
			if err := json.Unmarshal(m.Message, &r); err != nil {
				p.fatalError(err)
				return
			}
			err := p.handleReceivedRequest(r)
			if err != nil {
				p.fatalError(fmt.Errorf("failed to handle received request: %w", err))
				return
			}
		case response:
			if err := json.Unmarshal(m.Message, &res); err != nil {
				p.fatalError(err)
				return
			}
			err = p.handleReceivedResponse(res)
			if err != nil {
				p.fatalError(fmt.Errorf("failed to handle received response: %w", err))
				return
			}
		}
	}
}

func (p *Peer) handleReceivedUpdate(u ReceivedUpdate) error {
	switch u.Type {
	default:
		return p.updateHandler(u)
	}
}

func (p *Peer) handleReceivedRequest(r ReceivedRequest) error {
	var (
		res any
		err error
	)

	switch r.Type {
	default:
		res, err = p.requestHandler(r)
	}

	if err != nil {
		return err
	}

	return p.send(message{
		MessageType: response,
		Message: Response{
			RequestID: r.ID,
			Data:      res,
		},
	})
}

func (p *Peer) handleReceivedResponse(u ReceivedResponse) error {
	p.mu.RLock()
	resChan, ok := p.responseMap[u.RequestID]
	p.mu.RUnlock()
	if !ok {
		return ErrUnexpectedResponse
	}

	resChan <- u
	p.unregisterRequestID(u.RequestID)
	return nil
}
