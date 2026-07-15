// Package redis manages Sentinel's Redis client connection.
package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// Redis owns the Redis client used by Sentinel.
type Redis struct {
	Client *goredis.Client
}

// New connects to Redis, validates the connection, and returns a client owner.
func New(ctx context.Context, address, password string) (*Redis, error) {
	client := goredis.NewClient(&goredis.Options{Addr: address, Password: password})
	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping Redis: %w; close client: %v", err, closeErr)
		}
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return &Redis{Client: client}, nil
}

// Close releases Redis resources owned by the client.
func (r *Redis) Close() error {
	return r.Client.Close()
}
