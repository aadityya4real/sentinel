// Package database manages Sentinel's PostgreSQL connection and schema migrations.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database owns the PostgreSQL connection pool used by Sentinel.
type Database struct {
	Pool *pgxpool.Pool
}

// New connects to PostgreSQL, validates the connection, and returns a pool owner.
func New(ctx context.Context, connectionString string) (*Database, error) {
	cfg, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL configuration: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingContext); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// Close releases all PostgreSQL connections owned by the database.
func (d *Database) Close() {
	d.Pool.Close()
}
