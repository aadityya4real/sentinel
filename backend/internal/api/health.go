package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const apiVersion = "0.1.0"

// Pinger verifies that an infrastructure dependency is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthResponse reports the live status of Sentinel's infrastructure dependencies.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
	Uptime   string `json:"uptime"`
	Version  string `json:"version"`
}

// HealthHandler reports infrastructure health by pinging PostgreSQL and Redis.
type HealthHandler struct {
	database  Pinger
	redis     Pinger
	logger    *zap.Logger
	startedAt time.Time
}

// NewHealthHandler creates a handler that pings infrastructure dependencies.
func NewHealthHandler(database, redis Pinger, logger *zap.Logger) (*HealthHandler, error) {
	if database == nil {
		return nil, errors.New("database pinger is required")
	}
	if redis == nil {
		return nil, errors.New("redis pinger is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	return &HealthHandler{
		database:  database,
		redis:     redis,
		logger:    logger,
		startedAt: time.Now(),
	}, nil
}

// ServeHTTP handles GET /api/v1/health by pinging PostgreSQL and Redis.
func (h *HealthHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	dbStatus := pingDependency(ctx, h.database, h.logger, "database")
	redisStatus := pingDependency(ctx, h.redis, h.logger, "redis")

	status := "healthy"
	if dbStatus == "disconnected" || redisStatus == "disconnected" {
		status = "unhealthy"
	}

	response := HealthResponse{
		Status:   status,
		Database: dbStatus,
		Redis:    redisStatus,
		Uptime:   time.Since(h.startedAt).Round(time.Second).String(),
		Version:  apiVersion,
	}

	code := http.StatusOK
	if status == "unhealthy" {
		code = http.StatusServiceUnavailable
	}
	writeJSON(writer, code, response)
}

func pingDependency(ctx context.Context, pinger Pinger, logger *zap.Logger, name string) string {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := pinger.Ping(pingCtx); err != nil {
		logger.Warn("health check failed", zap.String("dependency", name), zap.Error(err))
		return "disconnected"
	}
	return "connected"
}
