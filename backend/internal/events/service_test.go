package events

import (
	"context"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/aadityya4real/sentinel/backend/internal/models"
)

type storeStub struct {
	event eventstore.NewEvent
	calls int
}

func (s *storeStub) Append(_ context.Context, event eventstore.NewEvent) (eventstore.Event, error) {
	s.event = event
	s.calls++
	return eventstore.Event{ID: 7}, nil
}

type cacheStub struct {
	event models.Event
	calls int
}

func (s *cacheStub) Store(_ context.Context, event models.Event) error {
	s.event = event
	s.calls++
	return nil
}

func TestCollectorCollectStoresAndCachesEvent(t *testing.T) {
	store := &storeStub{}
	cache := &cacheStub{}
	collector, err := NewCollector(store, cache)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	stored, err := collector.Collect(context.Background(), models.Event{Type: "infrastructure.cpu.changed", Hostname: "node-01", OccurredAt: time.Now(), Payload: []byte(`{"usage":90}`)})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if stored.ID != 7 || store.calls != 1 || cache.calls != 1 || store.event.Key == "" {
		t.Fatalf("event was not stored and cached correctly: %+v", store)
	}
}
