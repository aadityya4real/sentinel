package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/aadityya4real/sentinel/backend/internal/events"
	"github.com/aadityya4real/sentinel/backend/internal/models"
	"go.uber.org/zap"
)

// EventCollector receives and persists infrastructure events from Sentinel Agents.
type EventCollector interface {
	Collect(ctx context.Context, event models.Event) (eventstore.Event, error)
}

// EventsHandler receives infrastructure events from Sentinel Agents.
type EventsHandler struct {
	collector EventCollector
	logger    *zap.Logger
}

// NewEventsHandler creates an HTTP handler for Agent event ingestion.
func NewEventsHandler(collector EventCollector, logger *zap.Logger) (*EventsHandler, error) {
	if collector == nil {
		return nil, errors.New("event collector is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	return &EventsHandler{collector: collector, logger: logger}, nil
}

// ServeHTTP validates and persists a single Agent infrastructure event.
func (h *EventsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !acceptsJSON(request.Header.Get("Content-Type")) {
		writeError(writer, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return
	}
	if request.ContentLength > maxMetricsPayloadSize {
		writeError(writer, http.StatusRequestEntityTooLarge, "payload_too_large", errPayloadTooLarge.Error())
		return
	}
	var event models.Event
	if err := decodeJSON(writer, request, &event); err != nil {
		if errors.Is(err, errPayloadTooLarge) {
			writeError(writer, http.StatusRequestEntityTooLarge, "payload_too_large", err.Error())
			return
		}
		writeError(writer, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	stored, err := h.collector.Collect(request.Context(), event)
	if err != nil {
		var validationError *events.ValidationError
		if errors.As(err, &validationError) {
			writeError(writer, http.StatusUnprocessableEntity, "validation_failed", validationError.Error())
			return
		}
		h.logger.Error("collect infrastructure event", zap.Error(err), zap.String("hostname", event.Hostname), zap.String("event_type", event.Type))
		writeError(writer, http.StatusServiceUnavailable, "event_storage_unavailable", "event storage is temporarily unavailable")
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{"status": "accepted", "event_id": stored.ID})
}
