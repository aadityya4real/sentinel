package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

type stubPinger struct {
	err error
}

func (s *stubPinger) Ping(context.Context) error {
	return s.err
}

func TestHealthHandlerReturnsHealthyWhenAllDependenciesConnected(t *testing.T) {
	handler, err := NewHealthHandler(&stubPinger{}, &stubPinger{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewHealthHandler() error = %v", err)
	}

	rec := serveHealth(handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var body HealthResponse
	decodeJSONBody(t, rec, &body)
	if body.Status != "healthy" {
		t.Fatalf("status = %q, want %q", body.Status, "healthy")
	}
	if body.Database != "connected" {
		t.Fatalf("database = %q, want %q", body.Database, "connected")
	}
	if body.Redis != "connected" {
		t.Fatalf("redis = %q, want %q", body.Redis, "connected")
	}
	if body.Version != apiVersion {
		t.Fatalf("version = %q, want %q", body.Version, apiVersion)
	}
}

func TestHealthHandlerReturnsUnhealthyWhenDatabaseDown(t *testing.T) {
	handler, err := NewHealthHandler(&stubPinger{err: errors.New("connection refused")}, &stubPinger{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewHealthHandler() error = %v", err)
	}

	rec := serveHealth(handler)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body HealthResponse
	decodeJSONBody(t, rec, &body)
	if body.Status != "unhealthy" {
		t.Fatalf("status = %q, want %q", body.Status, "unhealthy")
	}
	if body.Database != "disconnected" {
		t.Fatalf("database = %q, want %q", body.Database, "disconnected")
	}
}

func TestHealthHandlerReturnsUnhealthyWhenRedisDown(t *testing.T) {
	handler, err := NewHealthHandler(&stubPinger{}, &stubPinger{err: errors.New("connection refused")}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewHealthHandler() error = %v", err)
	}

	rec := serveHealth(handler)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body HealthResponse
	decodeJSONBody(t, rec, &body)
	if body.Status != "unhealthy" {
		t.Fatalf("status = %q, want %q", body.Status, "unhealthy")
	}
	if body.Redis != "disconnected" {
		t.Fatalf("redis = %q, want %q", body.Redis, "disconnected")
	}
}

func TestNewHealthHandlerRejectsNilDependencies(t *testing.T) {
	logger := zap.NewNop()
	if _, err := NewHealthHandler(nil, &stubPinger{}, logger); err == nil {
		t.Fatal("expected error for nil database pinger")
	}
	if _, err := NewHealthHandler(&stubPinger{}, nil, logger); err == nil {
		t.Fatal("expected error for nil redis pinger")
	}
	if _, err := NewHealthHandler(&stubPinger{}, &stubPinger{}, nil); err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func serveHealth(handler *HealthHandler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeJSONBody(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}
