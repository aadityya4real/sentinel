package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

func TestMetricsCollectedEventKeyIsStableAcrossRetries(t *testing.T) {
	metrics := validMetrics()
	first, err := newMetricsCollectedEvent(metrics)
	if err != nil {
		t.Fatal(err)
	}
	second, err := newMetricsCollectedEvent(metrics)
	if err != nil {
		t.Fatal(err)
	}
	if first.Key == "" || first.Key != second.Key {
		t.Fatalf("retry keys differ: %q != %q", first.Key, second.Key)
	}
}

type memoryCache struct {
	stored int
	err    error
}

type memoryIngestionStore struct {
	stored  int
	err     error
	metrics agent.Metrics
	event   eventstore.NewEvent
}

func (s *memoryIngestionStore) StoreMetricAndEvent(_ context.Context, metrics agent.Metrics, event eventstore.NewEvent) error {
	s.stored++
	s.metrics = metrics
	s.event = event
	return s.err
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
	store := &memoryIngestionStore{}
	cache := &memoryCache{}
	broadcaster := &memoryBroadcaster{}
	service, err := NewService(store, cache, broadcaster)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := service.Record(context.Background(), validMetrics()); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if store.stored != 1 || cache.stored != 1 || broadcaster.published != 1 {
		t.Fatalf("stored transaction=%d cache=%d broadcast=%d, want 1 each", store.stored, cache.stored, broadcaster.published)
	}
	if store.event.Type != "infrastructure.metrics.collected" || store.event.Key == "" {
		t.Fatalf("event = %+v, want keyed infrastructure metric event", store.event)
	}
}

func TestServiceRecordRejectsInvalidMetricsBeforeStorage(t *testing.T) {
	store := &memoryIngestionStore{}
	cache := &memoryCache{}
	broadcaster := &memoryBroadcaster{}
	service, err := NewService(store, cache, broadcaster)
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
	if store.stored != 0 || cache.stored != 0 || broadcaster.published != 0 {
		t.Fatalf("invalid metrics must not be stored, got transaction=%d cache=%d broadcast=%d", store.stored, cache.stored, broadcaster.published)
	}
}

func TestServiceRecordStopsAfterPostgresTransactionFails(t *testing.T) {
	store := &memoryIngestionStore{err: errors.New("transaction failed")}
	cache := &memoryCache{}
	broadcaster := &memoryBroadcaster{}
	service, err := NewService(store, cache, broadcaster)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := service.Record(context.Background(), validMetrics()); err == nil {
		t.Fatal("Record() error = nil, want transaction error")
	}
	if store.stored != 1 || cache.stored != 0 || broadcaster.published != 0 {
		t.Fatalf("after transaction failure: transaction=%d cache=%d broadcast=%d", store.stored, cache.stored, broadcaster.published)
	}
}

func TestServiceRecordDoesNotFailWhenRedisCacheFails(t *testing.T) {
	store := &memoryIngestionStore{}
	cache := &memoryCache{err: errors.New("redis unavailable")}
	broadcaster := &memoryBroadcaster{}
	service, err := NewService(store, cache, broadcaster)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := service.Record(context.Background(), validMetrics()); err != nil {
		t.Fatalf("Record() error = %v, want successful durable ingestion", err)
	}
	if store.stored != 1 || cache.stored != 1 || broadcaster.published != 1 {
		t.Fatalf("stored transaction=%d cache=%d broadcast=%d, want best-effort projections attempted", store.stored, cache.stored, broadcaster.published)
	}
}

func TestServiceRecordDoesNotFailWhenBroadcastFails(t *testing.T) {
	store := &memoryIngestionStore{}
	cache := &memoryCache{}
	broadcaster := &memoryBroadcaster{err: errors.New("websocket unavailable")}
	service, err := NewService(store, cache, broadcaster)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := service.Record(context.Background(), validMetrics()); err != nil {
		t.Fatalf("Record() error = %v, want successful ingestion", err)
	}
	if store.stored != 1 || cache.stored != 1 || broadcaster.published != 1 {
		t.Fatalf("stored transaction=%d cache=%d broadcast=%d, want 1 each", store.stored, cache.stored, broadcaster.published)
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
