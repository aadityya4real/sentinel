package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/collector"
	"go.uber.org/zap"
)

type stubHealthReader struct {
	keys    []string
	metrics map[string]agent.Metrics
	err     error
}

func (s *stubHealthReader) ScanKeys(_ context.Context) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.keys, nil
}

func (s *stubHealthReader) Get(_ context.Context, hostname string) (agent.Metrics, error) {
	if s.err != nil {
		return agent.Metrics{}, s.err
	}
	m, ok := s.metrics[hostname]
	if !ok {
		return agent.Metrics{}, collector.ErrCacheMiss
	}
	return m, nil
}

func activeMetrics() agent.Metrics {
	return agent.Metrics{
		CPUUsagePercent: 25.5,
		Memory:          agent.MemoryUsage{UsedPercent: 50.0},
		Hostname:        "node-01",
		OS:              "linux",
		Timestamp:       time.Now().UTC(),
	}
}

func staleMetrics() agent.Metrics {
	return agent.Metrics{
		CPUUsagePercent: 80.0,
		Memory:          agent.MemoryUsage{UsedPercent: 90.0},
		Hostname:        "node-02",
		OS:              "linux",
		Timestamp:       time.Now().UTC().Add(-10 * time.Minute),
	}
}

func TestHealthHandlerReturnsHealthyWhenHostsActive(t *testing.T) {
	logger := zap.NewNop()
	now := time.Now().UTC()
	reader := &stubHealthReader{
		keys: []string{"sentinel:metrics:latest:node-01"},
		metrics: map[string]agent.Metrics{
			"node-01": activeMetrics(),
		},
	}

	handler, err := NewHealthHandler(reader, logger)
	if err != nil {
		t.Fatalf("NewHealthHandler() error = %v", err)
	}
	handler.now = func() time.Time { return now }

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body HealthResponse
	decodeJSONBody(t, rec, &body)
	if body.Status != "healthy" {
		t.Fatalf("status = %q, want %q", body.Status, "healthy")
	}
	if len(body.Hosts) != 1 || body.Hosts[0].Hostname != "node-01" || body.Hosts[0].Status != "healthy" {
		t.Fatalf("hosts = %+v, want [node-01 healthy]", body.Hosts)
	}
}

func TestHealthHandlerReturnsDegradedWhenHostsStale(t *testing.T) {
	logger := zap.NewNop()
	reader := &stubHealthReader{
		keys: []string{"sentinel:metrics:latest:node-02"},
		metrics: map[string]agent.Metrics{
			"node-02": staleMetrics(),
		},
	}

	handler, err := NewHealthHandler(reader, logger)
	if err != nil {
		t.Fatalf("NewHealthHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body HealthResponse
	decodeJSONBody(t, rec, &body)
	if body.Status != "degraded" {
		t.Fatalf("status = %q, want %q", body.Status, "degraded")
	}
	if len(body.Hosts) != 1 || body.Hosts[0].Status != "degraded" {
		t.Fatalf("hosts = %+v, want [node-02 degraded]", body.Hosts)
	}
}

func TestHealthHandlerReturnsUnhealthyWhenNoHosts(t *testing.T) {
	logger := zap.NewNop()
	reader := &stubHealthReader{keys: nil}

	handler, err := NewHealthHandler(reader, logger)
	if err != nil {
		t.Fatalf("NewHealthHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body HealthResponse
	decodeJSONBody(t, rec, &body)
	if body.Status != "unhealthy" {
		t.Fatalf("status = %q, want %q", body.Status, "unhealthy")
	}
	if len(body.Hosts) != 0 {
		t.Fatalf("hosts = %+v, want empty", body.Hosts)
	}
}

func TestHealthHandlerReturnsServiceUnavailableOnScanError(t *testing.T) {
	logger := zap.NewNop()
	reader := &stubHealthReader{err: errors.New("redis down")}

	handler, err := NewHealthHandler(reader, logger)
	if err != nil {
		t.Fatalf("NewHealthHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestNewHealthHandlerRejectsNilDependencies(t *testing.T) {
	logger := zap.NewNop()
	reader := &stubHealthReader{}

	if _, err := NewHealthHandler(nil, logger); err == nil {
		t.Fatal("expected error for nil cache")
	}
	if _, err := NewHealthHandler(reader, nil); err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func decodeJSONBody(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}
