// Package models contains shared Sentinel API data models.
package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Event is an infrastructure event submitted by a Sentinel Agent.
type Event struct {
	ID         string          `json:"id,omitempty"`
	Type       string          `json:"type"`
	Hostname   string          `json:"hostname"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

// Validate verifies that an event has all data required for durable ingestion.
func (e Event) Validate() error {
	if len(e.ID) > 512 {
		return fmt.Errorf("id must not exceed 512 characters")
	}
	if strings.TrimSpace(e.Type) == "" || len(e.Type) > 255 {
		return fmt.Errorf("type must be between 1 and 255 characters")
	}
	if strings.TrimSpace(e.Hostname) == "" || len(e.Hostname) > 255 {
		return fmt.Errorf("hostname must be between 1 and 255 characters")
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("occurred_at is required")
	}
	if len(e.Payload) == 0 || len(e.Payload) > 1<<20 || !json.Valid(e.Payload) {
		return fmt.Errorf("payload must contain valid JSON up to 1048576 bytes")
	}

	return nil
}
