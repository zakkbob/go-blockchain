package gossip

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
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

type Response struct {
	RequestID int `json:"request_id"`
	Data      any `json:"data"`
}

type ReceivedResponse struct {
	RequestID int             `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

type Peer struct {
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

	p := &Peer{
		conn:           conn,
		updateHandler:  updateHandler,
		requestHandler: requestHandler,
		responseMap:    map[int]chan ReceivedResponse{},
	}

	go p.handle()

	return p, nil
}

func (p *Peer) Update(updateType string, data any) error {
	return p.send(message{
		MessageType: update,
		Message: Update{
			Type: updateType,
			Data: data,
		},
	})
}

func (p *Peer) Request(ctx context.Context, requestType string, data any) (ReceivedResponse, error) {
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
	case res := <-resChan:
		return res, nil
	case <-ctx.Done():
		p.unregisterRequestID(id)
		return ReceivedResponse{}, ctx.Err()
	}
}

// not proper yet
func (p *Peer) fatalError(err error) {
	p.conn.Close()
	p.closeErr = err
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
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			p.fatalError(err)
			continue
		}

		switch m.MessageType {
		case update:
			if err := json.Unmarshal(m.Message, &u); err != nil {
				p.fatalError(err)
				return
			}
			err = p.handleReceivedUpdate(u)
			if err != nil {
				p.fatalError(err)
			}
		case request:
			if err := json.Unmarshal(m.Message, &r); err != nil {
				p.fatalError(err)
				continue
			}
			err = p.handleReceivedUpdate(u)
			if err != nil {
				p.handleReceivedRequest(r)
			}
		case response:
			if err := json.Unmarshal(m.Message, &res); err != nil {
				p.fatalError(err)
				continue
			}
			err = p.handleReceivedUpdate(u)
			if err != nil {
				p.handleReceivedResponse(res)
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
		return errors.New("Received response does not match a request")
	}

	resChan <- u
	p.unregisterRequestID(u.RequestID)
	return nil
}
