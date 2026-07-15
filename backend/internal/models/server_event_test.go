package models

import (
	"testing"
)

func TestServerEventValidateCreate(t *testing.T) {
	event := ServerEvent{ServerID: 1, EventType: "cpu.threshold.exceeded", Severity: "high", Payload: []byte(`{"usage":95}`)}
	if err := event.ValidateCreate(); err != nil {
		t.Fatalf("ValidateCreate() error = %v", err)
	}
}

func TestServerEventValidateCreateRejectsInvalidPayload(t *testing.T) {
	event := ServerEvent{ServerID: 1, EventType: "cpu.threshold.exceeded", Severity: "high", Payload: []byte(`invalid`)}
	if err := event.ValidateCreate(); err == nil {
		t.Fatal("ValidateCreate() error = nil, want payload validation error")
	}
}
