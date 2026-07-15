package api

import (
	"bytes"
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

type recorderStub struct {
	recorded agent.Metrics
	err      error
}

func (s *recorderStub) Record(_ context.Context, metrics agent.Metrics) error {
	s.recorded = metrics
	return s.err
}

func TestMetricsHandlerAcceptsValidPayload(t *testing.T) {
	recorder := &recorderStub{}
	handler, err := NewMetricsHandler(recorder, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMetricsHandler() error = %v", err)
	}

	response := executeRequest(t, handler, http.MethodPost, validPayload(t), "application/json")
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
	if recorder.recorded.Hostname != "node-01" {
		t.Fatalf("hostname = %q, want node-01", recorder.recorded.Hostname)
	}
}

func TestMetricsHandlerRejectsUnknownFields(t *testing.T) {
	handler, err := NewMetricsHandler(&recorderStub{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMetricsHandler() error = %v", err)
	}

	response := executeRequest(t, handler, http.MethodPost, []byte(`{"unexpected": true}`), "application/json")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestMetricsHandlerReturnsValidationFailure(t *testing.T) {
	handler, err := NewMetricsHandler(&recorderStub{err: &collector.ValidationError{Field: "hostname", Message: "is required"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMetricsHandler() error = %v", err)
	}

	response := executeRequest(t, handler, http.MethodPost, validPayload(t), "application/json")
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
}

func TestMetricsHandlerReturnsStorageUnavailable(t *testing.T) {
	handler, err := NewMetricsHandler(&recorderStub{err: errors.New("Redis unavailable")}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMetricsHandler() error = %v", err)
	}

	response := executeRequest(t, handler, http.MethodPost, validPayload(t), "application/json")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestMetricsHandlerRejectsOversizedPayload(t *testing.T) {
	handler, err := NewMetricsHandler(&recorderStub{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMetricsHandler() error = %v", err)
	}

	response := executeRequest(t, handler, http.MethodPost, bytes.Repeat([]byte("a"), maxMetricsPayloadSize+1), "application/json")
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func executeRequest(t *testing.T, handler http.Handler, method string, body []byte, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, "/v1/metrics", bytes.NewReader(body))
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func validPayload(t *testing.T) []byte {
	t.Helper()
	payload, err := json.Marshal(agent.Metrics{
		CPUUsagePercent: 20,
		Memory: agent.MemoryUsage{
			TotalBytes:     1024,
			UsedBytes:      512,
			AvailableBytes: 512,
			UsedPercent:    50,
		},
		Disks: []agent.DiskUsage{{
			Path:        "/",
			Filesystem:  "ext4",
			TotalBytes:  1024,
			UsedBytes:   512,
			UsedPercent: 50,
		}},
		Hostname:      "node-01",
		OS:            "linux",
		UptimeSeconds: 3600,
		Timestamp:     time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return payload
}
