// Package collector validates and stores infrastructure metric events.
package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
)

const maxInt64 = uint64(^uint64(0) >> 1)

// MetricsRepository persists immutable infrastructure metric events.
type MetricsRepository interface {
	Store(ctx context.Context, metrics agent.Metrics) error
}

// ErrCacheMiss is returned when a requested entry does not exist in the cache.
var ErrCacheMiss = errors.New("cache miss")

// LatestMetricsCache stores and retrieves the most recent metric event for each host.
type LatestMetricsCache interface {
	Store(ctx context.Context, metrics agent.Metrics) error
	Get(ctx context.Context, hostname string) (agent.Metrics, error)
}

// EventAppender appends immutable infrastructure events to the event store.
type EventAppender interface {
	Append(ctx context.Context, event eventstore.NewEvent) (eventstore.Event, error)
}

// Recorder records validated infrastructure metric events.
type Recorder interface {
	Record(ctx context.Context, metrics agent.Metrics) error
}

// Service coordinates validation and storage of metrics received from agents.
type Service struct {
	repository MetricsRepository
	events     EventAppender
	cache      LatestMetricsCache
}

// NewService creates a metric collection service from the supplied storage dependencies.
func NewService(repository MetricsRepository, events EventAppender, cache LatestMetricsCache) (*Service, error) {
	if repository == nil {
		return nil, errors.New("metrics repository is required")
	}
	if events == nil {
		return nil, errors.New("event appender is required")
	}
	if cache == nil {
		return nil, errors.New("latest metrics cache is required")
	}
	return &Service{repository: repository, events: events, cache: cache}, nil
}

// Record validates a metric event, persists it, and refreshes the latest host snapshot.
func (s *Service) Record(ctx context.Context, metrics agent.Metrics) error {
	if err := Validate(metrics); err != nil {
		return err
	}
	if err := s.repository.Store(ctx, metrics); err != nil {
		return fmt.Errorf("store metric event: %w", err)
	}
	event, err := newMetricsCollectedEvent(metrics)
	if err != nil {
		return fmt.Errorf("create metric event: %w", err)
	}
	if _, err := s.events.Append(ctx, event); err != nil {
		return fmt.Errorf("append metric event: %w", err)
	}
	if err := s.cache.Store(ctx, metrics); err != nil {
		return fmt.Errorf("cache latest metric event: %w", err)
	}
	return nil
}

func newMetricsCollectedEvent(metrics agent.Metrics) (eventstore.NewEvent, error) {
	payload, err := json.Marshal(metrics)
	if err != nil {
		return eventstore.NewEvent{}, fmt.Errorf("marshal metric event payload: %w", err)
	}

	return eventstore.NewEvent{
		Key:         "metrics:" + metrics.Hostname + ":" + metrics.Timestamp.UTC().Format(time.RFC3339Nano),
		Type:        "infrastructure.metrics.collected",
		SubjectType: "host",
		SubjectID:   metrics.Hostname,
		OccurredAt:  metrics.Timestamp.UTC(),
		Payload:     payload,
	}, nil
}

// ValidationError identifies an invalid field in a submitted metrics payload.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns a client-safe validation error message.
func (e *ValidationError) Error() string {
	return e.Field + " " + e.Message
}

// Validate verifies that a metrics payload can be safely persisted and represented by Sentinel.
func Validate(metrics agent.Metrics) error {
	if value := strings.TrimSpace(metrics.Hostname); value == "" || len(value) > 255 {
		return &ValidationError{Field: "hostname", Message: "must be between 1 and 255 characters"}
	}
	if value := strings.TrimSpace(metrics.OS); value == "" || len(value) > 255 {
		return &ValidationError{Field: "os", Message: "must be between 1 and 255 characters"}
	}
	if metrics.Timestamp.IsZero() {
		return &ValidationError{Field: "timestamp", Message: "is required"}
	}
	if metrics.Timestamp.After(time.Now().Add(5 * time.Minute)) {
		return &ValidationError{Field: "timestamp", Message: "must not be more than five minutes in the future"}
	}
	if err := validatePercent("cpu_usage_percent", metrics.CPUUsagePercent); err != nil {
		return err
	}
	if metrics.UptimeSeconds > maxInt64 {
		return &ValidationError{Field: "uptime_seconds", Message: "exceeds supported range"}
	}
	if err := validateMemory(metrics.Memory); err != nil {
		return err
	}
	if len(metrics.Disks) == 0 {
		return &ValidationError{Field: "disks", Message: "must contain at least one mounted filesystem"}
	}

	paths := make(map[string]struct{}, len(metrics.Disks))
	for index, usage := range metrics.Disks {
		field := fmt.Sprintf("disks[%d]", index)
		if strings.TrimSpace(usage.Path) == "" || len(usage.Path) > 1024 {
			return &ValidationError{Field: field + ".path", Message: "must be between 1 and 1024 characters"}
		}
		if _, exists := paths[usage.Path]; exists {
			return &ValidationError{Field: field + ".path", Message: "must be unique"}
		}
		paths[usage.Path] = struct{}{}
		if usage.TotalBytes == 0 || usage.TotalBytes > maxInt64 || usage.UsedBytes > usage.TotalBytes {
			return &ValidationError{Field: field + ".total_bytes", Message: "must be within the supported range"}
		}
		if err := validatePercent(field+".used_percent", usage.UsedPercent); err != nil {
			return err
		}
	}
	return nil
}

func validateMemory(memory agent.MemoryUsage) error {
	if memory.TotalBytes == 0 || memory.TotalBytes > maxInt64 || memory.UsedBytes > memory.TotalBytes || memory.AvailableBytes > memory.TotalBytes {
		return &ValidationError{Field: "memory", Message: "contains invalid byte values"}
	}
	return validatePercent("memory.used_percent", memory.UsedPercent)
}

func validatePercent(field string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 100 {
		return &ValidationError{Field: field, Message: "must be a number between 0 and 100"}
	}
	return nil
}
