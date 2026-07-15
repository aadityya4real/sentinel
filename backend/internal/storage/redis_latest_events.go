package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/models"
	goredis "github.com/redis/go-redis/v9"
)

// RedisLatestEventCache stores the most recently received event for each host.
type RedisLatestEventCache struct{ client *goredis.Client }

// NewRedisLatestEventCache creates a latest-event cache backed by Redis.
func NewRedisLatestEventCache(client *goredis.Client) (*RedisLatestEventCache, error) {
	if client == nil {
		return nil, fmt.Errorf("Redis client is required")
	}
	return &RedisLatestEventCache{client: client}, nil
}

// Store serializes and stores the latest event for the submitting host.
func (c *RedisLatestEventCache) Store(ctx context.Context, event models.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal latest event: %w", err)
	}
	if err := c.client.Set(ctx, "sentinel:events:latest:"+url.PathEscape(event.Hostname), payload, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("set latest event: %w", err)
	}
	return nil
}
