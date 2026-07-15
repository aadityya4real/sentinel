package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	goredis "github.com/redis/go-redis/v9"
)

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
	if err := c.client.Set(ctx, latestMetricsKey(metrics.Hostname), payload, 0).Err(); err != nil {
		return fmt.Errorf("set latest metrics: %w", err)
	}
	return nil
}

func latestMetricsKey(hostname string) string {
	return "sentinel:metrics:latest:" + url.PathEscape(hostname)
}
