package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

type memoryRepository struct {
	stored int
	err    error
}

func (r *memoryRepository) Store(context.Context, agent.Metrics) error {
	r.stored++
	return r.err
}

type memoryCache struct {
	stored int
	err    error
}

type memoryEventAppender struct {
	stored int
	err    error
	event  eventstore.NewEvent
}

func (a *memoryEventAppender) Append(_ context.Context, event eventstore.NewEvent) (eventstore.Event, error) {
	a.stored++
	a.event = event
	return eventstore.Event{}, a.err
}

func (c *memoryCache) Store(context.Context, agent.Metrics) error {
	c.stored++
	return c.err
}

func (c *memoryCache) Get(_ context.Context, _ string) (agent.Metrics, error) {
	return agent.Metrics{}, ErrCacheMiss
}

type memoryBroadcaster struct {
	published int
	err       error
}

func (b *memoryBroadcaster) Publish(context.Context, agent.Metrics) error {
	b.published++
	return b.err
}

func TestServiceRecordStoresValidatedMetrics(t *testing.T) {
	repository := &memoryRepository{}
	events := &memoryEventAppender{}
	cache := &memoryCache{}
	broadcaster := &memoryBroadcaster{}
	service, err := NewService(repository, events, cache, broadcaster)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := service.Record(context.Background(), validMetrics()); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if repository.stored != 1 || events.stored != 1 || cache.stored != 1 || broadcaster.published != 1 {
		t.Fatalf("stored repository=%d events=%d cache=%d broadcast=%d, want 1 each", repository.stored, events.stored, cache.stored, broadcaster.published)
	}
	if events.event.Type != "infrastructure.metrics.collected" {
		t.Fatalf("event type = %q, want infrastructure.metrics.collected", events.event.Type)
	}
}

func TestServiceRecordRejectsInvalidMetricsBeforeStorage(t *testing.T) {
	repository := &memoryRepository{}
	events := &memoryEventAppender{}
	cache := &memoryCache{}
	broadcaster := &memoryBroadcaster{}
	service, err := NewService(repository, events, cache, broadcaster)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	metrics := validMetrics()
	metrics.Hostname = ""
	err = service.Record(context.Background(), metrics)
	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("Record() error = %v, want ValidationError", err)
	}
	if repository.stored != 0 || events.stored != 0 || cache.stored != 0 || broadcaster.published != 0 {
		t.Fatalf("invalid metrics must not be stored, got repository=%d events=%d cache=%d broadcast=%d", repository.stored, events.stored, cache.stored, broadcaster.published)
	}
}

func validMetrics() agent.Metrics {
	return agent.Metrics{
		CPUUsagePercent: 25.5,
		Memory: agent.MemoryUsage{
			TotalBytes:     16 * 1024,
			UsedBytes:      8 * 1024,
			AvailableBytes: 8 * 1024,
			UsedPercent:    50,
		},
		Disks: []agent.DiskUsage{{
			Path:        "/",
			Filesystem:  "ext4",
			TotalBytes:  100 * 1024,
			UsedBytes:   50 * 1024,
			UsedPercent: 50,
		}},
		Hostname:      "node-01",
		OS:            "linux",
		UptimeSeconds: 3600,
		Timestamp:     time.Now().UTC(),
	}
}
