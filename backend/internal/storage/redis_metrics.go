package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/collector"
	goredis "github.com/redis/go-redis/v9"
)

const metricsCacheTTL = 5 * time.Minute

// RedisLatestMetricsCache stores the latest metric snapshot for each host in Redis.
type RedisLatestMetricsCache struct {
	client *goredis.Client
}

// NewRedisLatestMetricsCache creates a latest-metrics cache backed by Redis.
func NewRedisLatestMetricsCache(client *goredis.Client) (*RedisLatestMetricsCache, error) {
	if client == nil {
		return nil, fmt.Errorf("Redis client is required")
	}
	return &RedisLatestMetricsCache{client: client}, nil
}

// Store serializes and stores the latest metric snapshot for the event's hostname.
func (c *RedisLatestMetricsCache) Store(ctx context.Context, metrics agent.Metrics) error {
	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal latest metrics: %w", err)
	}
	if err := c.client.Set(ctx, latestMetricsKey(metrics.Hostname), payload, metricsCacheTTL).Err(); err != nil {
		return fmt.Errorf("set latest metrics: %w", err)
	}
	return nil
}

// Get retrieves the latest cached metric snapshot for a host.
func (c *RedisLatestMetricsCache) Get(ctx context.Context, hostname string) (agent.Metrics, error) {
	cmd := c.client.Get(ctx, latestMetricsKey(hostname))
	raw, err := cmd.Bytes()
	if err != nil {
		if err == goredis.Nil {
			return agent.Metrics{}, collector.ErrCacheMiss
		}
		return agent.Metrics{}, fmt.Errorf("get latest metrics: %w", err)
	}
	var metrics agent.Metrics
	if err := json.Unmarshal(raw, &metrics); err != nil {
		return agent.Metrics{}, fmt.Errorf("unmarshal latest metrics: %w", err)
	}
	return metrics, nil
}

// ScanKeys returns all cached hostnames using the metrics key pattern.
func (c *RedisLatestMetricsCache) ScanKeys(ctx context.Context) ([]string, error) {
	var keys []string
	var cursor uint64
	const match = "sentinel:metrics:latest:*"
	const count = 100
	for {
		var batch []string
		var err error
		batch, cursor, err = c.client.Scan(ctx, cursor, match, count).Result()
		if err != nil {
			return nil, fmt.Errorf("scan metrics keys: %w", err)
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

func latestMetricsKey(hostname string) string {
	return "sentinel:metrics:latest:" + url.PathEscape(hostname)
}
