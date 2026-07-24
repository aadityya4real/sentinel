package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10 // must be < pongWait
	sendBufferSize = 64
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a single connected WebSocket subscriber.
type Client struct {
	hub   *Hub
	conn  *websocket.Conn
	send  chan []byte
	log   *zap.Logger
}

// newClient wires a connection to the hub and returns an unregistered client.
func newClient(hub *Hub, conn *websocket.Conn, logger *zap.Logger) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, sendBufferSize),
		log:  logger,
	}
}

// readPump drains incoming messages (clients do not send meaningful payloads)
// and enforces the pong deadline so dead connections are reaped.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				c.log.Debug("websocket read error", zap.Error(err))
			}
			return
		}
	}
}

// writePump flushes queued broadcast messages to the client and sends periodic
// pings to keep the connection alive. It exits when the send channel is closed
// (either by the hub on shutdown or by a slow-client drop).
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
