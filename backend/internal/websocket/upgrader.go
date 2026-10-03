package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// NewUpgrader creates an upgrader constrained to the configured browser origins.
func NewUpgrader(allowedOrigins []string) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     func(request *http.Request) bool { return OriginAllowed(request, allowedOrigins) },
	}
}

// OriginAllowed permits non-browser clients without an Origin header and otherwise requires an exact allow-list match.
func OriginAllowed(request *http.Request, allowedOrigins []string) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// UpgradeHTTP promotes the HTTP connection, registers the resulting client
// with the hub, and spawns its read/write goroutines.
func UpgradeHTTP(hub *Hub, logger *zap.Logger, allowedOrigins []string, writer http.ResponseWriter, request *http.Request) {
	upgrader := NewUpgrader(allowedOrigins)
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		logger.Debug("websocket upgrade failed", zap.Error(err))
		return
	}

	client := newClient(hub, conn, logger)
	hub.register <- client

	go client.writePump()
	go client.readPump()
}
