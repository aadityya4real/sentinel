package ai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

type latestEventReaderStub struct {
	events []eventstore.Event
	filter eventstore.Filter
}

func (s *latestEventReaderStub) ListLatest(_ context.Context, filter eventstore.Filter) ([]eventstore.Event, error) {
	s.filter = filter
	if len(s.events) > filter.Limit {
		return s.events[:filter.Limit], nil
	}
	return s.events, nil
}

type completionStub struct{ prompt string }

func (s *completionStub) Complete(_ context.Context, _, prompt string) (string, error) {
	s.prompt = prompt
	return `{"summary":"ok","severity":"low","confidence":0.8}`, nil
}

func TestAnalyzeSelectsNewestEventsAndPresentsThemChronologically(t *testing.T) {
	at := time.Date(2026, 10, 4, 14, 0, 0, 0, time.UTC)
	reader := &latestEventReaderStub{events: []eventstore.Event{
		{ID: 5, OccurredAt: at.Add(4 * time.Minute)}, {ID: 4, OccurredAt: at.Add(3 * time.Minute)}, {ID: 3, OccurredAt: at.Add(2 * time.Minute)}, {ID: 2, OccurredAt: at.Add(time.Minute)}, {ID: 1, OccurredAt: at},
	}}
	client := &completionStub{}
	service, err := NewService(reader, client)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Analyze(context.Background(), Request{Hostname: "node-1", From: at, To: at.Add(5 * time.Minute), EventLimit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if reader.filter.Limit != 2 {
		t.Fatalf("limit=%d, want 2", reader.filter.Limit)
	}
	start := strings.Index(client.prompt, "Event data follows")
	var selected []eventstore.Event
	if err := json.Unmarshal([]byte(client.prompt[strings.Index(client.prompt, "\n")+1:]), &selected); err != nil {
		t.Fatalf("decode prompt: %v", err)
	}
	if len(selected) != 2 || selected[0].ID != 4 || selected[1].ID != 5 {
		t.Fatalf("prompt events=%v, want chronological IDs 4,5 (marker %d)", selected, start)
	}
}
