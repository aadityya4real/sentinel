package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ServerEvent is a relational infrastructure event persisted for a registered server.
type ServerEvent struct {
	ID        int64           `json:"id"`
	ServerID  int64           `json:"server_id"`
	EventType string          `json:"event_type"`
	Severity  string          `json:"severity"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

// ValidateCreate verifies that an event can be safely stored in the relational event table.
func (e ServerEvent) ValidateCreate() error {
	if e.ServerID < 1 {
		return fmt.Errorf("server_id must be positive")
	}
	if strings.TrimSpace(e.EventType) == "" || len(e.EventType) > 255 {
		return fmt.Errorf("event_type must be between 1 and 255 characters")
	}
	if strings.TrimSpace(e.Severity) == "" || len(e.Severity) > 255 {
		return fmt.Errorf("severity must be between 1 and 255 characters")
	}
	if len(e.Payload) == 0 || !json.Valid(e.Payload) {
		return fmt.Errorf("payload must contain valid JSON")
	}

	return nil
}
