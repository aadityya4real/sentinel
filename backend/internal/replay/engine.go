// Package replay reconstructs chronological infrastructure timelines from the event store.
package replay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

const maxReplayLimit = 999

// Request identifies a bounded host timeline to replay.
type Request struct {
	Hostname string
	From     time.Time
	To       time.Time
	Limit    int
	Cursor   string
}

// Timeline is an ordered page of host events suitable for a time-machine replay.
type Timeline struct {
	Hostname   string             `json:"hostname"`
	From       time.Time          `json:"from"`
	To         time.Time          `json:"to"`
	Events     []eventstore.Event `json:"events"`
	Limit      int                `json:"limit"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

// Reader replays bounded infrastructure timelines.
type Reader interface {
	Replay(ctx context.Context, request Request) (Timeline, error)
}

// Engine reads immutable events and produces ordered host timelines.
type Engine struct {
	store eventstore.Store
}

// NewEngine creates a replay engine backed by the supplied event store.
func NewEngine(store eventstore.Store) (*Engine, error) {
	if store == nil {
		return nil, errors.New("event store is required")
	}

	return &Engine{store: store}, nil
}

// Replay returns one chronological page of a host's infrastructure event timeline.
func (e *Engine) Replay(ctx context.Context, request Request) (Timeline, error) {
	if err := validateRequest(request); err != nil {
		return Timeline{}, err
	}

	filter := eventstore.Filter{
		SubjectType: "host",
		SubjectID:   request.Hostname,
		From:        request.From,
		To:          request.To,
		Limit:       request.Limit + 1,
	}
	if request.Cursor != "" {
		cursor, err := decodeCursor(request.Cursor)
		if err != nil {
			return Timeline{}, &ValidationError{Message: "cursor is invalid"}
		}
		filter.AfterAt = cursor.OccurredAt
		filter.AfterID = cursor.ID
	}

	events, err := e.store.List(ctx, filter)
	if err != nil {
		return Timeline{}, fmt.Errorf("read replay events: %w", err)
	}

	timeline := Timeline{
		Hostname: request.Hostname,
		From:     request.From,
		To:       request.To,
		Events:   events,
		Limit:    request.Limit,
	}
	if len(events) > request.Limit {
		timeline.Events = events[:request.Limit]
		timeline.NextCursor, err = encodeCursor(timeline.Events[len(timeline.Events)-1])
		if err != nil {
			return Timeline{}, fmt.Errorf("encode replay cursor: %w", err)
		}
	}

	return timeline, nil
}

// ValidationError identifies an invalid replay request parameter.
type ValidationError struct {
	Message string
}

// Error returns a client-safe validation error message.
func (e *ValidationError) Error() string {
	return e.Message
}

type cursor struct {
	OccurredAt time.Time `json:"occurred_at"`
	ID         int64     `json:"id"`
}

func validateRequest(request Request) error {
	if hostname := strings.TrimSpace(request.Hostname); hostname == "" || len(hostname) > 255 {
		return &ValidationError{Message: "hostname must be between 1 and 255 characters"}
	}
	if request.From.IsZero() || request.To.IsZero() || request.From.After(request.To) {
		return &ValidationError{Message: "from must be before or equal to to"}
	}
	if request.Limit < 1 || request.Limit > maxReplayLimit {
		return &ValidationError{Message: fmt.Sprintf("limit must be an integer between 1 and %d", maxReplayLimit)}
	}

	return nil
}

func encodeCursor(event eventstore.Event) (string, error) {
	data, err := json.Marshal(cursor{OccurredAt: event.OccurredAt.UTC(), ID: event.ID})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeCursor(value string) (cursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return cursor{}, err
	}

	var decoded cursor
	if err := json.Unmarshal(data, &decoded); err != nil {
		return cursor{}, err
	}
	if decoded.OccurredAt.IsZero() || decoded.ID < 1 {
		return cursor{}, errors.New("cursor fields are invalid")
	}

	return decoded, nil
}
