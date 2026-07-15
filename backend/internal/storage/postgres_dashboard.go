package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/dashboard"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLDashboardRepository serves dashboard read queries from PostgreSQL.
type PostgreSQLDashboardRepository struct {
	pool *pgxpool.Pool
}

// NewPostgreSQLDashboardRepository creates a dashboard repository backed by the supplied pool.
func NewPostgreSQLDashboardRepository(pool *pgxpool.Pool) (*PostgreSQLDashboardRepository, error) {
	if pool == nil {
		return nil, fmt.Errorf("PostgreSQL pool is required")
	}

	return &PostgreSQLDashboardRepository{pool: pool}, nil
}

// Overview returns fleet-level aggregates calculated from each host's latest metric event.
func (r *PostgreSQLDashboardRepository) Overview(ctx context.Context, activeSince time.Time) (dashboard.Overview, error) {
	var overview dashboard.Overview
	err := r.pool.QueryRow(ctx, `
		WITH latest_metrics AS (
			SELECT DISTINCT ON (hostname)
				hostname, collected_at, cpu_usage_percent, memory_used_percent
			FROM infrastructure_metrics
			ORDER BY hostname, collected_at DESC
		)
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE collected_at >= $1),
			COALESCE(AVG(cpu_usage_percent), 0),
			COALESCE(AVG(memory_used_percent), 0),
			MAX(collected_at)
		FROM latest_metrics`, activeSince).Scan(
		&overview.TotalHosts,
		&overview.ActiveHosts,
		&overview.AverageCPUUsagePercent,
		&overview.AverageMemoryUsagePercent,
		&overview.LatestMetricAt,
	)
	if err != nil {
		return dashboard.Overview{}, fmt.Errorf("query dashboard overview: %w", err)
	}

	return overview, nil
}

// LatestHosts returns each host's latest metric event, ordered by collection time.
func (r *PostgreSQLDashboardRepository) LatestHosts(ctx context.Context, limit int) ([]agent.Metrics, error) {
	rows, err := r.pool.Query(ctx, `
		WITH latest_metrics AS (
			SELECT DISTINCT ON (hostname)
				hostname, operating_system, uptime_seconds, collected_at, cpu_usage_percent,
				memory_total_bytes, memory_used_bytes, memory_available_bytes, memory_used_percent, disks
			FROM infrastructure_metrics
			ORDER BY hostname, collected_at DESC
		)
		SELECT hostname, operating_system, uptime_seconds, collected_at, cpu_usage_percent,
			memory_total_bytes, memory_used_bytes, memory_available_bytes, memory_used_percent, disks
		FROM latest_metrics
		ORDER BY collected_at DESC, hostname ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query latest host metrics: %w", err)
	}
	defer rows.Close()

	return collectMetrics(rows)
}

// History returns metric events for one host in ascending collection-time order.
func (r *PostgreSQLDashboardRepository) History(ctx context.Context, hostname string, from, to time.Time, limit int) ([]agent.Metrics, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT hostname, operating_system, uptime_seconds, collected_at, cpu_usage_percent,
			memory_total_bytes, memory_used_bytes, memory_available_bytes, memory_used_percent, disks
		FROM infrastructure_metrics
		WHERE hostname = $1 AND collected_at >= $2 AND collected_at <= $3
		ORDER BY collected_at ASC
		LIMIT $4`, hostname, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("query host metric history: %w", err)
	}
	defer rows.Close()

	return collectMetrics(rows)
}

func collectMetrics(rows pgx.Rows) ([]agent.Metrics, error) {
	metrics := make([]agent.Metrics, 0)
	for rows.Next() {
		metric, err := scanMetric(rows)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate metric rows: %w", err)
	}

	return metrics, nil
}

func scanMetric(row pgx.Row) (agent.Metrics, error) {
	var metric agent.Metrics
	var disks []byte
	var uptimeSeconds int64
	var totalBytes, usedBytes, availableBytes int64
	if err := row.Scan(
		&metric.Hostname,
		&metric.OS,
		&uptimeSeconds,
		&metric.Timestamp,
		&metric.CPUUsagePercent,
		&totalBytes,
		&usedBytes,
		&availableBytes,
		&metric.Memory.UsedPercent,
		&disks,
	); err != nil {
		return agent.Metrics{}, fmt.Errorf("scan metric row: %w", err)
	}
	if err := json.Unmarshal(disks, &metric.Disks); err != nil {
		return agent.Metrics{}, fmt.Errorf("decode disk usage: %w", err)
	}

	metric.UptimeSeconds = uint64(uptimeSeconds)
	metric.Memory.TotalBytes = uint64(totalBytes)
	metric.Memory.UsedBytes = uint64(usedBytes)
	metric.Memory.AvailableBytes = uint64(availableBytes)
	return metric, nil
}
