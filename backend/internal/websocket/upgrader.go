package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// upgrader promotes HTTP requests to WebSocket connections. Origin is not
// constrained in this default; production deployments should configure
// CheckOrigin against an allow-list when serving browsers directly.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// UpgradeHTTP promotes the HTTP connection, registers the resulting client
// with the hub, and spawns its read/write goroutines.
func UpgradeHTTP(hub *Hub, logger *zap.Logger, writer http.ResponseWriter, request *http.Request) {
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
