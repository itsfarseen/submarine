package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// Base Types

type RPC struct {
	conn      *websocket.Conn
	idCounter atomic.Uint64
	mu        sync.RWMutex
	pending   map[uint64]chan *json.RawMessage
	ctx       context.Context
	cancel    context.CancelFunc
}

type RpcRequest struct {
	ID      uint64 `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type RpcResponse struct {
	ID     uint64          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *RpcError       `json:"error"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RpcError) Error() string {
	return e.Message
}

// Pending Request

type PendingRequest struct {
	ch chan *json.RawMessage
}

func (p *PendingRequest) AsString() (string, error) {
	var res string
	err := p.As(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (p *PendingRequest) RawMessage() (*json.RawMessage, error) {
	rawResult, ok := <-p.ch
	if !ok || rawResult == nil {
		return nil, errors.New("request failed or connection closed")
	}
	return rawResult, nil
}

func (p *PendingRequest) As(value any) error {
	rawResult, err := p.RawMessage()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(*rawResult, value); err != nil {
		return err
	}
	return nil
}

func NewRPC(url string) (*RPC, error) {
	ctx, cancel := context.WithCancel(context.Background())
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		cancel()
		return nil, err
	}

	client := &RPC{
		conn:    conn,
		pending: make(map[uint64]chan *json.RawMessage),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Start the background loop to read messages
	go client.readLoop()

	return client, nil
}

func (r *RPC) readLoop() {
	defer r.conn.Close()
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
			_, message, err := r.conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				// Close all pending channels on error
				r.mu.Lock()
				for id, ch := range r.pending {
					close(ch)
					delete(r.pending, id)
				}
				r.mu.Unlock()
				return
			}

			var resp RpcResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				log.Printf("unmarshal error: %v", err)
				continue
			}

			if resp.Error != nil {
				log.Printf("RPC Error: %s", resp.Error.Error())
			}

			r.mu.RLock()
			ch, ok := r.pending[resp.ID]
			r.mu.RUnlock()

			if ok {
				ch <- &resp.Result
				// Clean up the map
				r.mu.Lock()
				delete(r.pending, resp.ID)
				close(ch)
				r.mu.Unlock()
			}
		}
	}
}

func (r *RPC) Send(method string, params []any) *PendingRequest {
	id := r.idCounter.Add(1)
	request := RpcRequest{
		ID:      id,
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}

	jsonReq, err := json.Marshal(request)
	if err != nil {
		// This should not happen with valid inputs
		log.Fatalf("failed to marshal request: %v", err)
	}

	respChan := make(chan *json.RawMessage, 1)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.pending[id] = respChan

	if err := r.conn.WriteMessage(websocket.TextMessage, jsonReq); err != nil {
		log.Printf("write error: %v", err)
		delete(r.pending, id)
		close(respChan)
	}

	return &PendingRequest{ch: respChan}
}

func (r *RPC) Close() {
	r.cancel()
	r.conn.Close()

	r.mu.Lock()
	for id, ch := range r.pending {
		close(ch)
		delete(r.pending, id)
	}
	r.mu.Unlock()
}
