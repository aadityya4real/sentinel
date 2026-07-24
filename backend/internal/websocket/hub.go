// Package websocket provides a thread-safe broadcast hub for live metric streaming.
package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"go.uber.org/zap"
)

const (
	broadcastBufferSize = 256
	registerBufferSize  = 16
)

// Hub fans out published metrics to every connected client. It is safe for
// concurrent use by any number of publishers and subscribers.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	mu         sync.RWMutex
	clients    map[*Client]struct{}
	clientCount atomic.Int64

	log     *zap.Logger
	closed  atomic.Bool
}

// NewHub creates a broadcast hub. Call Run to start the fan-out goroutine.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		register:   make(chan *Client, registerBufferSize),
		unregister: make(chan *Client, registerBufferSize),
		broadcast:  make(chan []byte, broadcastBufferSize),
		clients:    make(map[*Client]struct{}),
		log:        logger,
	}
}

// Run processes register, unregister, and broadcast events until ctx is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()
			h.clientCount.Store(int64(len(h.clients)))
			h.log.Debug("websocket client registered", zap.Int64("clients", h.clientCount.Load()))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.clientCount.Store(int64(len(h.clients)))
			h.log.Debug("websocket client unregistered", zap.Int64("clients", h.clientCount.Load()))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full — drop this message for this client
					// rather than blocking the whole fan-out. Slow clients get
					// eventually reaped by the heartbeat/ping check.
				}
			}
			h.mu.RUnlock()

		case <-ctx.Done():
			h.log.Info("websocket hub shutting down")
			return
		}
	}
}

// Publish serializes and broadcasts a metric snapshot to all subscribers.
// It is non-blocking: if the broadcast buffer is full the message is dropped
// so that the caller (the collector) is never stalled by slow consumers.
func (h *Hub) Publish(ctx context.Context, metrics agent.Metrics) error {
	if h.closed.Load() {
		return nil
	}
	payload, err := json.Marshal(struct {
		Type    string        `json:"type"`
		Payload agent.Metrics `json:"payload"`
	}{Type: "metrics", Payload: metrics})
	if err != nil {
		return err
	}
	select {
	case h.broadcast <- payload:
	case <-ctx.Done():
	default:
		h.log.Debug("broadcast buffer full, dropping metric", zap.String("hostname", metrics.Hostname))
	}
	return nil
}

// ClientCount returns the current number of connected clients.
func (h *Hub) ClientCount() int { return int(h.clientCount.Load()) }

// Close shuts down the hub. Subsequent Publish calls are no-ops.
func (h *Hub) Close() {
	if !h.closed.CompareAndSwap(false, true) {
		return
	}
	h.mu.Lock()
	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
	h.clients = make(map[*Client]struct{})
	h.mu.Unlock()
	h.clientCount.Store(0)
}
