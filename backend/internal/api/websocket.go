package api

import (
	"errors"
	"net/http"

	"github.com/aadityya4real/sentinel/backend/internal/websocket"
	"go.uber.org/zap"
)

// WebsocketHandler exposes the live metric stream over WebSocket.
type WebsocketHandler struct {
	hub    *websocket.Hub
	logger *zap.Logger
}

// NewWebsocketHandler creates a handler that upgrades requests onto the metric hub.
func NewWebsocketHandler(hub *websocket.Hub, logger *zap.Logger) (*WebsocketHandler, error) {
	if hub == nil {
		return nil, errors.New("websocket hub is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	return &WebsocketHandler{hub: hub, logger: logger}, nil
}

// Metrics upgrades an HTTP request to a WebSocket subscription on /ws/v1/metrics.
func (h *WebsocketHandler) Metrics(writer http.ResponseWriter, request *http.Request) {
	websocket.UpgradeHTTP(h.hub, h.logger, writer, request)
}
