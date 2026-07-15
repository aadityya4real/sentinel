package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aadityya4real/sentinel/backend/internal/events"
	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/aadityya4real/sentinel/backend/internal/models"
	"go.uber.org/zap"
)

type eventStoreStub struct{}

func (eventStoreStub) Append(context.Context, eventstore.NewEvent) (eventstore.Event, error) {
	return eventstore.Event{ID: 1}, nil
}

type latestEventStub struct{}

func (latestEventStub) Store(context.Context, models.Event) error { return nil }

func TestEventsHandlerAcceptsEvent(t *testing.T) {
	collector, err := events.NewCollector(eventStoreStub{}, latestEventStub{})
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	handler, err := NewEventsHandler(collector, zap.NewNop())
	if err != nil {
		t.Fatalf("NewEventsHandler() error = %v", err)
	}
	body := []byte(`{"type":"infrastructure.cpu.changed","hostname":"node-01","occurred_at":"2026-07-16T12:00:00Z","payload":{"usage":90}}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
}

func TestEventsHandlerRejectsInvalidEvent(t *testing.T) {
	collector, _ := events.NewCollector(eventStoreStub{}, latestEventStub{})
	handler, _ := NewEventsHandler(collector, zap.NewNop())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(`{"type":"","hostname":"node-01","occurred_at":"2026-07-16T12:00:00Z","payload":{}}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
}
