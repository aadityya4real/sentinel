package replay

import (
	"context"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

type eventStoreStub struct {
	events []eventstore.Event
	filter eventstore.Filter
}

func (s *eventStoreStub) Append(context.Context, eventstore.NewEvent) (eventstore.Event, error) {
	return eventstore.Event{}, nil
}

func (s *eventStoreStub) List(_ context.Context, filter eventstore.Filter) ([]eventstore.Event, error) {
	s.filter = filter
	return s.events, nil
}

func (s *eventStoreStub) Latest(context.Context, eventstore.Filter) (eventstore.Event, bool, error) {
	return eventstore.Event{}, false, nil
}

func TestEngineReplayPaginatesChronologicalEvents(t *testing.T) {
	occurredAt := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &eventStoreStub{events: []eventstore.Event{
		{ID: 1, OccurredAt: occurredAt},
		{ID: 2, OccurredAt: occurredAt},
		{ID: 3, OccurredAt: occurredAt.Add(time.Second)},
	}}
	engine, err := NewEngine(store)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	timeline, err := engine.Replay(context.Background(), Request{
		Hostname: "node-01",
		From:     occurredAt.Add(-time.Minute),
		To:       occurredAt.Add(time.Minute),
		Limit:    2,
	})
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if len(timeline.Events) != 2 || timeline.NextCursor == "" {
		t.Fatalf("events=%d next_cursor=%q, want two events and a cursor", len(timeline.Events), timeline.NextCursor)
	}
	if store.filter.SubjectType != "host" || store.filter.SubjectID != "node-01" || store.filter.Limit != 3 {
		t.Fatalf("filter = %+v, want host node-01 with limit 3", store.filter)
	}

	_, err = engine.Replay(context.Background(), Request{
		Hostname: "node-01",
		From:     occurredAt.Add(-time.Minute),
		To:       occurredAt.Add(time.Minute),
		Limit:    2,
		Cursor:   timeline.NextCursor,
	})
	if err != nil {
		t.Fatalf("Replay() with cursor error = %v", err)
	}
	if !store.filter.AfterAt.Equal(occurredAt) || store.filter.AfterID != 2 {
		t.Fatalf("cursor filter = %+v, want timestamp %s and ID 2", store.filter, occurredAt)
	}
}

func TestEngineReplayRejectsInvalidCursor(t *testing.T) {
	engine, err := NewEngine(&eventStoreStub{})
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	_, err = engine.Replay(context.Background(), Request{
		Hostname: "node-01",
		From:     time.Now().Add(-time.Minute),
		To:       time.Now(),
		Limit:    1,
		Cursor:   "invalid",
	})
	if err == nil {
		t.Fatal("Replay() error = nil, want invalid cursor error")
	}
}
