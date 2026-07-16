package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/dashboard"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type dashboardReaderStub struct {
	overview    dashboard.Overview
	hosts       dashboard.HostsPage
	history     dashboard.History
	err         error
	historyHost string
}

func (s *dashboardReaderStub) GetOverview(context.Context) (dashboard.Overview, error) {
	return s.overview, s.err
}

func (s *dashboardReaderStub) GetHosts(context.Context, int) (dashboard.HostsPage, error) {
	return s.hosts, s.err
}

func (s *dashboardReaderStub) GetHistory(_ context.Context, hostname string, _ time.Time, _ time.Time, _ int) (dashboard.History, error) {
	s.historyHost = hostname
	return s.history, s.err
}

func TestDashboardOverviewReturnsData(t *testing.T) {
	handler, err := NewDashboardHandler(&dashboardReaderStub{overview: dashboard.Overview{TotalHosts: 2}}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewDashboardHandler() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/overview", nil)
	response := httptest.NewRecorder()
	handler.Overview(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestDashboardHostsRejectsInvalidLimit(t *testing.T) {
	handler, err := NewDashboardHandler(&dashboardReaderStub{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewDashboardHandler() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/hosts?limit=251", nil)
	response := httptest.NewRecorder()
	handler.Hosts(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestDashboardHistoryReturnsBadRequestForInvalidRange(t *testing.T) {
	handler, err := NewDashboardHandler(&dashboardReaderStub{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewDashboardHandler() error = %v", err)
	}
	router := chi.NewRouter()
	router.Get("/api/v1/dashboard/hosts/{hostname}/metrics", handler.History)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/hosts/node-01/metrics?from=invalid", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestDashboardReturnsServiceUnavailable(t *testing.T) {
	handler, err := NewDashboardHandler(&dashboardReaderStub{err: errors.New("database unavailable")}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewDashboardHandler() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/overview", nil)
	response := httptest.NewRecorder()
	handler.Overview(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
