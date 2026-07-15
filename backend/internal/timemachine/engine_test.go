package timemachine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

type eventStoreStub struct {
	event  eventstore.Event
	found  bool
	err    error
	filter eventstore.Filter
}

func (s *eventStoreStub) Append(context.Context, eventstore.NewEvent) (eventstore.Event, error) {
	return eventstore.Event{}, nil
}

func (s *eventStoreStub) List(context.Context, eventstore.Filter) ([]eventstore.Event, error) {
	return nil, nil
}

func (s *eventStoreStub) Latest(_ context.Context, filter eventstore.Filter) (eventstore.Event, bool, error) {
	s.filter = filter
	return s.event, s.found, s.err
}

func TestEngineSnapshotReconstructsLatestMetrics(t *testing.T) {
	occurredAt := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &eventStoreStub{
		found: true,
		event: eventstore.Event{
			ID:         42,
			OccurredAt: occurredAt,
			Payload:    []byte(`{"hostname":"node-01","timestamp":"2026-07-15T12:00:00Z"}`),
		},
	}
	engine, err := NewEngine(store)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	requestedAt := occurredAt.Add(time.Hour)
	snapshot, err := engine.Snapshot(context.Background(), Request{Hostname: "node-01", At: requestedAt})
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snapshot.EventID != 42 || !snapshot.ObservedAt.Equal(occurredAt) || snapshot.Metrics.Hostname != "node-01" {
		t.Fatalf("snapshot = %+v, want reconstructed node-01 event", snapshot)
	}
	if store.filter.Type != metricCollectedEventType || store.filter.SubjectID != "node-01" || !store.filter.To.Equal(requestedAt) {
		t.Fatalf("filter = %+v, want metric event for node-01", store.filter)
	}
}

func TestEngineSnapshotReturnsNotFoundWhenNoEventExists(t *testing.T) {
	engine, err := NewEngine(&eventStoreStub{found: false})
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	_, err = engine.Snapshot(context.Background(), Request{Hostname: "node-01", At: time.Now()})
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("Snapshot() error = %v, want NotFoundError", err)
	}
}
