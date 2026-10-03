package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/aadityya4real/sentinel/backend/internal/events"
	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/aadityya4real/sentinel/backend/internal/models"
	"go.uber.org/zap"
)

type eventStoreStub struct {
	events     []eventstore.Event
	lastFilter eventstore.Filter
}

func (eventStoreStub) Append(context.Context, eventstore.NewEvent) (eventstore.Event, error) {
	return eventstore.Event{ID: 1}, nil
}

func (eventStoreStub) List(context.Context, eventstore.Filter) ([]eventstore.Event, error) {
	return []eventstore.Event{}, nil
}
func (s *eventStoreStub) ListLatest(_ context.Context, filter eventstore.Filter) ([]eventstore.Event, error) {
	s.lastFilter = filter
	result := make([]eventstore.Event, 0, filter.Limit)
	for _, event := range s.events {
		if !filter.BeforeAt.IsZero() && (event.OccurredAt.After(filter.BeforeAt) || (event.OccurredAt.Equal(filter.BeforeAt) && event.ID >= filter.BeforeID)) {
			continue
		}
		result = append(result, event)
		if len(result) == filter.Limit {
			break
		}
	}
	return result, nil
}

type latestEventStub struct{}

func (latestEventStub) Store(context.Context, models.Event) error { return nil }

func TestEventsHandlerAcceptsEvent(t *testing.T) {
	collector, err := events.NewCollector(&eventStoreStub{}, latestEventStub{})
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
	collector, _ := events.NewCollector(&eventStoreStub{}, latestEventStub{})
	handler, _ := NewEventsHandler(collector, zap.NewNop())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(`{"type":"","hostname":"node-01","occurred_at":"2026-07-16T12:00:00Z","payload":{}}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
}

func TestEventsListCursorPagesNewestFirstWithoutEqualTimestampGaps(t *testing.T) {
	at := time.Now().UTC().Truncate(time.Second)
	store := &eventStoreStub{events: []eventstore.Event{
		{ID: 5, OccurredAt: at.Add(2 * time.Second)}, {ID: 4, OccurredAt: at.Add(time.Second)},
		{ID: 3, OccurredAt: at}, {ID: 2, OccurredAt: at}, {ID: 1, OccurredAt: at},
	}}
	collector, err := events.NewCollector(store, latestEventStub{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewEventsHandler(collector, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	var collected []int64
	url := "/api/v1/events?limit=2"
	for page := 0; page < 3; page++ {
		response := httptest.NewRecorder()
		handler.List(response, httptest.NewRequest(http.MethodGet, url, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("page %d status=%d body=%s", page, response.Code, response.Body.String())
		}
		var result EventList
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		for _, event := range result.Events {
			collected = append(collected, event.ID)
		}
		if page < 2 {
			if result.NextCursor == "" {
				t.Fatalf("page %d missing next_cursor", page)
			}
			url = "/api/v1/events?limit=2&cursor=" + result.NextCursor
		} else if result.NextCursor != "" {
			t.Fatalf("last page has cursor %q", result.NextCursor)
		}
	}
	if fmt.Sprint(collected) != "[5 4 3 2 1]" {
		t.Fatalf("IDs=%v, expected complete newest-first pagination", collected)
	}
	if store.lastFilter.BeforeID != 2 || !store.lastFilter.BeforeAt.Equal(at) {
		t.Fatalf("last cursor filter=%+v, want (timestamp,ID)=(%s,2)", store.lastFilter, at)
	}
}
