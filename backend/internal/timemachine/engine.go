// Package timemachine reconstructs host state at a requested point in time.
package timemachine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

const metricCollectedEventType = "infrastructure.metrics.collected"

// Request identifies the host and point in time to reconstruct.
type Request struct {
	Hostname string
	At       time.Time
}

// Snapshot contains the reconstructed host state effective at the requested time.
type Snapshot struct {
	Hostname    string        `json:"hostname"`
	RequestedAt time.Time     `json:"requested_at"`
	ObservedAt  time.Time     `json:"observed_at"`
	EventID     int64         `json:"event_id"`
	Metrics     agent.Metrics `json:"metrics"`
}

// Reader reconstructs point-in-time host state.
type Reader interface {
	Snapshot(ctx context.Context, request Request) (Snapshot, error)
}

// Engine reconstructs host snapshots from immutable metric events.
type Engine struct {
	store eventstore.Store
}

// NewEngine creates a Time Machine engine backed by the supplied event store.
func NewEngine(store eventstore.Store) (*Engine, error) {
	if store == nil {
		return nil, errors.New("event store is required")
	}

	return &Engine{store: store}, nil
}

// Snapshot returns the latest complete metric snapshot at or before the requested time.
func (e *Engine) Snapshot(ctx context.Context, request Request) (Snapshot, error) {
	if err := validateRequest(request); err != nil {
		return Snapshot{}, err
	}

	event, found, err := e.store.Latest(ctx, eventstore.Filter{
		Type:        metricCollectedEventType,
		SubjectType: "host",
		SubjectID:   request.Hostname,
		To:          request.At,
	})
	if err != nil {
		return Snapshot{}, fmt.Errorf("read point-in-time event: %w", err)
	}
	if !found {
		return Snapshot{}, &NotFoundError{Hostname: request.Hostname, At: request.At}
	}

	var metrics agent.Metrics
	if err := json.Unmarshal(event.Payload, &metrics); err != nil {
		return Snapshot{}, fmt.Errorf("decode metric event payload: %w", err)
	}
	if metrics.Hostname != request.Hostname {
		return Snapshot{}, fmt.Errorf("metric event hostname does not match event subject")
	}

	return Snapshot{
		Hostname:    request.Hostname,
		RequestedAt: request.At,
		ObservedAt:  event.OccurredAt,
		EventID:     event.ID,
		Metrics:     metrics,
	}, nil
}

// ValidationError identifies an invalid Time Machine request parameter.
type ValidationError struct {
	Message string
}

// Error returns a client-safe validation error message.
func (e *ValidationError) Error() string {
	return e.Message
}

// NotFoundError indicates that no host snapshot existed at the requested time.
type NotFoundError struct {
	Hostname string
	At       time.Time
}

// Error returns a client-safe missing-snapshot message.
func (e *NotFoundError) Error() string {
	return "no host snapshot exists at the requested time"
}

func validateRequest(request Request) error {
	if hostname := strings.TrimSpace(request.Hostname); hostname == "" || len(hostname) > 255 {
		return &ValidationError{Message: "hostname must be between 1 and 255 characters"}
	}
	if request.At.IsZero() {
		return &ValidationError{Message: "at is required"}
	}

	return nil
}
