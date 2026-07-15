// Package events implements infrastructure-event ingestion.
package events

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/aadityya4real/sentinel/backend/internal/models"
)

// Store appends immutable infrastructure events.
type Store interface {
	Append(ctx context.Context, event eventstore.NewEvent) (eventstore.Event, error)
}

// LatestStateStore stores the most recent event submitted by each host.
type LatestStateStore interface {
	Store(ctx context.Context, event models.Event) error
}

// Collector receives validated Agent events and coordinates durable storage.
type Collector struct {
	store Store
	cache LatestStateStore
}

// NewCollector creates an event collector with durable and latest-state storage.
func NewCollector(store Store, cache LatestStateStore) (*Collector, error) {
	if store == nil {
		return nil, errors.New("event store is required")
	}
	if cache == nil {
		return nil, errors.New("latest state store is required")
	}
	return &Collector{store: store, cache: cache}, nil
}

// Collect validates, persists, and caches an Agent infrastructure event.
func (c *Collector) Collect(ctx context.Context, event models.Event) (eventstore.Event, error) {
	if err := event.Validate(); err != nil {
		return eventstore.Event{}, &ValidationError{Message: err.Error()}
	}
	event.OccurredAt = event.OccurredAt.UTC()
	key := event.ID
	if key == "" {
		key = eventKey(event)
	}
	stored, err := c.store.Append(ctx, eventstore.NewEvent{
		Key: key, Type: event.Type, SubjectType: "host", SubjectID: event.Hostname, OccurredAt: event.OccurredAt, Payload: event.Payload,
	})
	if err != nil {
		return eventstore.Event{}, fmt.Errorf("append event: %w", err)
	}
	if err := c.cache.Store(ctx, event); err != nil {
		return eventstore.Event{}, fmt.Errorf("cache latest state: %w", err)
	}
	return stored, nil
}

// ValidationError identifies an event payload rejected before storage.
type ValidationError struct{ Message string }

// Error returns a client-safe validation error message.
func (e *ValidationError) Error() string { return e.Message }

func eventKey(event models.Event) string {
	digest := sha256.Sum256(append([]byte(event.Type+"\x00"+event.Hostname+"\x00"+event.OccurredAt.Format(time.RFC3339Nano)+"\x00"), event.Payload...))
	return "agent:" + hex.EncodeToString(digest[:])
}
