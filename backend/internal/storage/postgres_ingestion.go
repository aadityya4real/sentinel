package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLIngestionStore commits each metric together with its immutable event.
type PostgreSQLIngestionStore struct {
	pool *pgxpool.Pool
}

// NewPostgreSQLIngestionStore creates an ingestion store backed by PostgreSQL.
func NewPostgreSQLIngestionStore(pool *pgxpool.Pool) (*PostgreSQLIngestionStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("PostgreSQL pool is required")
	}
	return &PostgreSQLIngestionStore{pool: pool}, nil
}

// StoreMetricAndEvent atomically upserts a metric and inserts its corresponding event.
func (s *PostgreSQLIngestionStore) StoreMetricAndEvent(ctx context.Context, metrics agent.Metrics, event eventstore.NewEvent) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("validate metric event: %w", err)
	}
	disks, err := json.Marshal(metrics.Disks)
	if err != nil {
		return fmt.Errorf("marshal disk usage: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin metric ingestion transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO infrastructure_metrics (
			hostname, operating_system, uptime_seconds, collected_at, cpu_usage_percent,
			memory_total_bytes, memory_used_bytes, memory_available_bytes, memory_used_percent, disks
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (hostname, collected_at) DO UPDATE SET
			operating_system = EXCLUDED.operating_system,
			uptime_seconds = EXCLUDED.uptime_seconds,
			cpu_usage_percent = EXCLUDED.cpu_usage_percent,
			memory_total_bytes = EXCLUDED.memory_total_bytes,
			memory_used_bytes = EXCLUDED.memory_used_bytes,
			memory_available_bytes = EXCLUDED.memory_available_bytes,
			memory_used_percent = EXCLUDED.memory_used_percent,
			disks = EXCLUDED.disks`,
		metrics.Hostname,
		metrics.OS,
		int64(metrics.UptimeSeconds),
		metrics.Timestamp,
		metrics.CPUUsagePercent,
		int64(metrics.Memory.TotalBytes),
		int64(metrics.Memory.UsedBytes),
		int64(metrics.Memory.AvailableBytes),
		metrics.Memory.UsedPercent,
		disks,
	)
	if err != nil {
		return fmt.Errorf("upsert infrastructure metrics: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO infrastructure_events (
			event_key, event_type, subject_type, subject_id, occurred_at, payload
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (event_key) DO NOTHING`,
		event.Key,
		event.Type,
		event.SubjectType,
		event.SubjectID,
		event.OccurredAt,
		event.Payload,
	)
	if err != nil {
		return fmt.Errorf("insert infrastructure event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit metric ingestion transaction: %w", err)
	}
	return nil
}
