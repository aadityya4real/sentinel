package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/collector"
)

func TestMetricsEndpoint(t *testing.T) {
	endpoint, err := metricsEndpoint("")
	if err != nil {
		t.Fatalf("metricsEndpoint() error = %v", err)
	}
	if endpoint != "http://localhost:8080/api/v1/metrics" {
		t.Fatalf("endpoint = %q", endpoint)
	}

	endpoint, err = metricsEndpoint("https://sentinel.example/prefix/")
	if err != nil {
		t.Fatalf("metricsEndpoint() error = %v", err)
	}
	if endpoint != "https://sentinel.example/prefix/api/v1/metrics" {
		t.Fatalf("endpoint = %q", endpoint)
	}
	if _, err := metricsEndpoint("localhost:8080"); err == nil {
		t.Fatal("metricsEndpoint() accepted a URL without an HTTP scheme")
	}
	if _, err := metricsEndpoint("http://sentinel.example"); err == nil {
		t.Fatal("metricsEndpoint() accepted remote HTTP, which would expose the bearer token")
	}
	if endpoint, err := metricsEndpoint("http://127.0.0.1:8080"); err != nil || endpoint != "http://127.0.0.1:8080/api/v1/metrics" {
		t.Fatalf("metricsEndpoint() local HTTP = %q, %v", endpoint, err)
	}
}

func TestPostMetricsSendsValidatedJSON(t *testing.T) {
	metrics := validAgentMetrics()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/metrics" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Errorf("Content-Type = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-agent-token-32-characters-long" {
			t.Errorf("Authorization header = %q, want configured bearer token", got)
		}
		var received agent.Metrics
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if err := collector.Validate(received); err != nil {
			t.Errorf("API validation rejected agent payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	if err := postMetrics(context.Background(), server.Client(), server.URL+"/api/v1/metrics", "test-agent-token-32-characters-long", metrics); err != nil {
		t.Fatalf("postMetrics() error = %v", err)
	}
}

func TestCollectAndSendRetriesServerErrors(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	if err := collectAndSend(context.Background(), fixedCollector{metrics: validAgentMetrics()}, server.Client(), server.URL, "test-agent-token-32-characters-long"); err != nil {
		t.Fatalf("collectAndSend() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("request attempts = %d, want 2", attempts)
	}
}

func TestCollectAndSendDoesNotRetryClientErrors(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		http.Error(w, "invalid metrics", http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	err := collectAndSend(context.Background(), fixedCollector{metrics: validAgentMetrics()}, server.Client(), server.URL, "test-agent-token-32-characters-long")
	if err == nil {
		t.Fatal("collectAndSend() error = nil, want API error")
	}
	if attempts != 1 {
		t.Fatalf("request attempts = %d, want 1", attempts)
	}
}

type fixedCollector struct{ metrics agent.Metrics }

func (c fixedCollector) Collect(context.Context) (agent.Metrics, error) { return c.metrics, nil }

func validAgentMetrics() agent.Metrics {
	return agent.Metrics{
		CPUUsagePercent: 20,
		Memory: agent.MemoryUsage{
			TotalBytes: 1024, UsedBytes: 512, AvailableBytes: 512, UsedPercent: 50,
		},
		Disks: []agent.DiskUsage{{
			Path: "/", Filesystem: "testfs", TotalBytes: 1000, UsedBytes: 500, UsedPercent: 50,
		}},
		Hostname: "test-host", OS: "test-os", UptimeSeconds: 42, Timestamp: time.Now().UTC(),
	}
}
