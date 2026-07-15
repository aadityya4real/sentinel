// Package storage provides PostgreSQL and Redis persistence for Sentinel metrics.
package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLMetricsRepository stores infrastructure metric events in PostgreSQL.
type PostgreSQLMetricsRepository struct {
	pool *pgxpool.Pool
}

// NewPostgreSQLMetricsRepository creates a repository backed by the supplied pool.
func NewPostgreSQLMetricsRepository(pool *pgxpool.Pool) (*PostgreSQLMetricsRepository, error) {
	if pool == nil {
		return nil, fmt.Errorf("PostgreSQL pool is required")
	}
	return &PostgreSQLMetricsRepository{pool: pool}, nil
}

// Store inserts a metric event or updates an exact retry from the same host and collection time.
func (r *PostgreSQLMetricsRepository) Store(ctx context.Context, metrics agent.Metrics) error {
	disks, err := json.Marshal(metrics.Disks)
	if err != nil {
		return fmt.Errorf("marshal disk usage: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
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
	return nil
}
