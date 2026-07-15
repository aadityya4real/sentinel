package eventstore

import (
	"testing"
	"time"
)

func TestNewEventValidateAcceptsCompleteEvent(t *testing.T) {
	event := NewEvent{
		Key:         "metrics:node-01:2026-07-15T12:00:00Z",
		Type:        "infrastructure.metrics.collected",
		SubjectType: "host",
		SubjectID:   "node-01",
		OccurredAt:  time.Now().UTC(),
		Payload:     []byte(`{"hostname":"node-01"}`),
	}
	if err := event.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestNewEventValidateRejectsInvalidPayload(t *testing.T) {
	event := NewEvent{
		Key:         "event-1",
		Type:        "infrastructure.metrics.collected",
		SubjectType: "host",
		SubjectID:   "node-01",
		OccurredAt:  time.Now().UTC(),
		Payload:     []byte(`not-json`),
	}
	if err := event.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want invalid payload error")
	}
}
