package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"go.uber.org/zap"
)

const healthyWindow = 5 * time.Minute

// HealthReader reads cached infrastructure metrics for health checks.
type HealthReader interface {
	ScanKeys(ctx context.Context) ([]string, error)
	Get(ctx context.Context, hostname string) (agent.Metrics, error)
}

// HostHealth describes the health status of a single monitored host.
type HostHealth struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	CPU      string `json:"cpu"`
	Memory   string `json:"memory"`
	OS       string `json:"os"`
}

// HealthResponse contains the aggregate health status of the fleet.
type HealthResponse struct {
	Status  string       `json:"status"`
	Hosts   []HostHealth `json:"hosts"`
	Version string       `json:"version"`
}

// HealthHandler reports fleet health from the metrics cache.
type HealthHandler struct {
	cache  HealthReader
	logger *zap.Logger
	now    func() time.Time
}

// NewHealthHandler creates a handler that serves infrastructure health from the metrics cache.
func NewHealthHandler(cache HealthReader, logger *zap.Logger) (*HealthHandler, error) {
	if cache == nil {
		return nil, errors.New("health reader is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	return &HealthHandler{cache: cache, logger: logger, now: time.Now}, nil
}

// ServeHTTP handles GET /health by reading cached metrics and classifying host health.
func (h *HealthHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	response, err := h.buildResponse(ctx)
	if err != nil {
		h.logger.Error("build health response", zap.Error(err))
		writeError(writer, http.StatusServiceUnavailable, "health_unavailable", "health check failed")
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (h *HealthHandler) buildResponse(ctx context.Context) (HealthResponse, error) {
	keys, err := h.cache.ScanKeys(ctx)
	if err != nil {
		return HealthResponse{}, err
	}

	now := h.now().UTC()
	cutoff := now.Add(-healthyWindow)

	hosts := make([]HostHealth, 0, len(keys))
	allHealthy := true

	for _, key := range keys {
		hostname := hostnameFromKey(key)
		metrics, err := h.cache.Get(ctx, hostname)
		if err != nil {
			h.logger.Warn("skip host in health check", zap.String("key", key), zap.Error(err))
			continue
		}
		status := "healthy"
		if metrics.Timestamp.Before(cutoff) {
			status = "degraded"
			allHealthy = false
		}
		hosts = append(hosts, HostHealth{
			Hostname: hostname,
			Status:   status,
			CPU:      fmt.Sprintf("%.1f%%", metrics.CPUUsagePercent),
			Memory:   fmt.Sprintf("%.1f%%", metrics.Memory.UsedPercent),
			OS:       metrics.OS,
		})
	}

	fleetStatus := "unhealthy"
	if len(hosts) > 0 {
		fleetStatus = "degraded"
	}
	if allHealthy && len(hosts) > 0 {
		fleetStatus = "healthy"
	}

	return HealthResponse{Status: fleetStatus, Hosts: hosts, Version: "0.1.0"}, nil
}

func hostnameFromKey(key string) string {
	const prefix = "sentinel:metrics:latest:"
	if !strings.HasPrefix(key, prefix) {
		return key
	}
	encoded := strings.TrimPrefix(key, prefix)
	decoded, err := url.PathUnescape(encoded)
	if err != nil {
		return encoded
	}
	return decoded
}
