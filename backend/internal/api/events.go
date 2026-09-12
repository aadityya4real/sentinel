package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/aadityya4real/sentinel/backend/internal/events"
	"github.com/aadityya4real/sentinel/backend/internal/models"
	"go.uber.org/zap"
)

const (
	defaultEventsLimit = 50
	maxEventsLimit     = 500
	eventsDefaultBack  = 24 * time.Hour
	eventsMaxRange     = 30 * 24 * time.Hour
)

// EventList contains a page of infrastructure events.
type EventList struct {
	Events []eventstore.Event `json:"events"`
	Limit  int                `json:"limit"`
	From   *time.Time         `json:"from,omitempty"`
	To     *time.Time         `json:"to,omitempty"`
}

// EventCollector receives and persists infrastructure events from Sentinel Agents.
type EventCollector interface {
	Collect(ctx context.Context, event models.Event) (eventstore.Event, error)
	List(ctx context.Context, filter eventstore.Filter) ([]eventstore.Event, error)
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

// List returns a page of infrastructure events for dashboard and events pages.
func (h *EventsHandler) List(writer http.ResponseWriter, request *http.Request) {
	filter := eventstore.Filter{Limit: defaultEventsLimit}

	if subjectType := request.URL.Query().Get("subject_type"); subjectType != "" {
		filter.SubjectType = strings.TrimSpace(subjectType)
	}
	if subjectID := request.URL.Query().Get("subject_id"); subjectID != "" {
		filter.SubjectID = strings.TrimSpace(subjectID)
	}
	if eventType := request.URL.Query().Get("type"); eventType != "" {
		filter.Type = strings.TrimSpace(eventType)
	}

	fromStr := request.URL.Query().Get("from")
	toStr := request.URL.Query().Get("to")

	from, to, err := parseEventTimeRange(fromStr, toStr, eventsDefaultBack, eventsMaxRange)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_time_range", err.Error())
		return
	}
	filter.From = from
	filter.To = to

	if limitStr := request.URL.Query().Get("limit"); limitStr != "" {
		limit, parseErr := strconv.Atoi(limitStr)
		if parseErr != nil || limit < 1 || limit > maxEventsLimit {
			writeError(writer, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and "+strconv.Itoa(maxEventsLimit))
			return
		}
		filter.Limit = limit
	}

	evts, err := h.collector.List(request.Context(), filter)
	if err != nil {
		h.logger.Error("list events", zap.Error(err))
		writeError(writer, http.StatusServiceUnavailable, "events_unavailable", "event list is temporarily unavailable")
		return
	}

	writeJSON(writer, http.StatusOK, EventList{
		Events: evts,
		Limit:  filter.Limit,
		From:   &from,
		To:     &to,
	})
}

// parseEventTimeRange returns from/to times with sensible defaults.
func parseEventTimeRange(rawFrom, rawTo string, defaultBack, maximum time.Duration) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	from := now.Add(-defaultBack)
	to := now

	var err error
	if rawFrom != "" {
		from, err = time.Parse(time.RFC3339, rawFrom)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("from must be an RFC3339 timestamp")
		}
	}
	if rawTo != "" {
		to, err = time.Parse(time.RFC3339, rawTo)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("to must be an RFC3339 timestamp")
		}
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, errors.New("from must be before or equal to to")
	}
	if to.Sub(from) > maximum {
		return time.Time{}, time.Time{}, errors.New("time range must not exceed thirty days")
	}

	return from.UTC(), to.UTC(), nil
}
