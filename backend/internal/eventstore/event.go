// Package eventstore defines Sentinel's durable infrastructure event timeline.
package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const maxPayloadSize = 1 << 20

// Event is an immutable event stored in Sentinel's infrastructure timeline.
type Event struct {
	ID          int64           `json:"id"`
	Key         string          `json:"key"`
	Type        string          `json:"type"`
	SubjectType string          `json:"subject_type"`
	SubjectID   string          `json:"subject_id"`
	OccurredAt  time.Time       `json:"occurred_at"`
	RecordedAt  time.Time       `json:"recorded_at"`
	Payload     json.RawMessage `json:"payload"`
}

// NewEvent contains the fields required to append an immutable event.
type NewEvent struct {
	Key         string          `json:"key"`
	Type        string          `json:"type"`
	SubjectType string          `json:"subject_type"`
	SubjectID   string          `json:"subject_id"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Payload     json.RawMessage `json:"payload"`
}

// Filter restricts a chronological event-store read.
type Filter struct {
	Type        string
	SubjectType string
	SubjectID   string
	From        time.Time
	To          time.Time
	AfterAt     time.Time
	AfterID     int64
	Limit       int
}

// Store appends and reads immutable infrastructure events.
type Store interface {
	Append(ctx context.Context, event NewEvent) (Event, error)
	List(ctx context.Context, filter Filter) ([]Event, error)
	Latest(ctx context.Context, filter Filter) (Event, bool, error)
}

// Validate verifies that an event has a safe, complete representation for durable storage.
func (e NewEvent) Validate() error {
	if err := validateText("key", e.Key, 512); err != nil {
		return err
	}
	if err := validateText("type", e.Type, 255); err != nil {
		return err
	}
	if err := validateText("subject_type", e.SubjectType, 255); err != nil {
		return err
	}
	if err := validateText("subject_id", e.SubjectID, 512); err != nil {
		return err
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("occurred_at is required")
	}
	if len(e.Payload) == 0 || len(e.Payload) > maxPayloadSize || !json.Valid(e.Payload) {
		return fmt.Errorf("payload must contain valid JSON up to %d bytes", maxPayloadSize)
	}

	return nil
}

func validateText(field, value string, maximum int) error {
	if strings.TrimSpace(value) == "" || len(value) > maximum {
		return fmt.Errorf("%s must be between 1 and %d characters", field, maximum)
	}

	return nil
}
