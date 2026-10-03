package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"go.uber.org/zap"
)

type securityRecorder struct{}

func (securityRecorder) Record(context.Context, agent.Metrics) error { return nil }

func TestRouterProtectsIngestionAndAIEndpoints(t *testing.T) {
	const token = "sentinel-router-test-token-at-least-32-chars"
	metrics, err := NewMetricsHandler(securityRecorder{}, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Handlers{Metrics: metrics}, zap.NewNop(), RouterSecurity{APIToken: token, AllowedOrigins: []string{"http://localhost:3000"}})
	for _, path := range []string{"/api/v1/metrics", "/api/v1/events", "/api/v1/ai/incidents/analyze"} {
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("anonymous POST %s status=%d, want 401", path, response.Code)
		}
	}

	body := `{"cpu_usage_percent":20,"memory":{"total_bytes":1024,"used_bytes":512,"available_bytes":512,"used_percent":50},"disks":[{"path":"/","filesystem":"test","total_bytes":1000,"used_bytes":500,"used_percent":50}],"hostname":"node-1","os":"linux","uptime_seconds":42,"timestamp":"2026-10-04T12:00:00Z"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/metrics", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("authenticated metric request status=%d, body=%s", response.Code, response.Body.String())
	}
}
