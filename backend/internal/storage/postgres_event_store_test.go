package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

func TestEventListQueriesUseExplicitStableOrderAndCursor(t *testing.T) {
	stamp := time.Date(2026, 10, 4, 14, 30, 0, 0, time.UTC)
	filter := eventstore.Filter{Limit: 3, BeforeAt: stamp, BeforeID: 17}
	query, args := buildEventListQuery(filter, "DESC")
	if !strings.Contains(query, "(occurred_at, id) < ($1, $2)") || !strings.Contains(query, "ORDER BY occurred_at DESC, id DESC") {
		t.Fatalf("query = %s, missing exclusive stable descending cursor", query)
	}
	if len(args) != 3 || args[0] != stamp || args[1] != int64(17) || args[2] != 3 {
		t.Fatalf("args = %#v, want timestamp, id, limit", args)
	}
	chronological, _ := buildEventListQuery(eventstore.Filter{Limit: 2}, "ASC")
	if !strings.Contains(chronological, "ORDER BY occurred_at ASC, id ASC") {
		t.Fatalf("chronological query = %s", chronological)
	}
}

func TestLatestEventQueryHonorsInclusivePointInTime(t *testing.T) {
	at := time.Date(2026, 10, 4, 14, 30, 0, 0, time.UTC)
	query, args := buildLatestEventQuery(eventstore.Filter{Type: "metric", SubjectID: "node-1", To: at})
	if !strings.Contains(query, "occurred_at <= $3") || !strings.Contains(query, "ORDER BY occurred_at DESC, id DESC") {
		t.Fatalf("query = %s, expected inclusive point-in-time latest ordering", query)
	}
	if len(args) != 3 || args[2] != at {
		t.Fatalf("args = %#v, expected requested time bound", args)
	}
}
