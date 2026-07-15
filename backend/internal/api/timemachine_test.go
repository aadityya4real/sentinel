package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aadityya4real/sentinel/backend/internal/timemachine"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type timeMachineReaderStub struct {
	snapshot timemachine.Snapshot
	err      error
}

func (s *timeMachineReaderStub) Snapshot(context.Context, timemachine.Request) (timemachine.Snapshot, error) {
	return s.snapshot, s.err
}

func TestTimeMachineHandlerReturnsSnapshot(t *testing.T) {
	handler, err := NewTimeMachineHandler(&timeMachineReaderStub{snapshot: timemachine.Snapshot{Hostname: "node-01"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTimeMachineHandler() error = %v", err)
	}
	router := timeMachineRouter(handler)

	request := httptest.NewRequest(http.MethodGet, "/v1/time-machine/hosts/node-01?at=2026-07-15T12:00:00Z", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestTimeMachineHandlerReturnsNotFound(t *testing.T) {
	handler, err := NewTimeMachineHandler(&timeMachineReaderStub{err: &timemachine.NotFoundError{}}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTimeMachineHandler() error = %v", err)
	}
	router := timeMachineRouter(handler)

	request := httptest.NewRequest(http.MethodGet, "/v1/time-machine/hosts/node-01?at=2026-07-15T12:00:00Z", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestTimeMachineHandlerRejectsMissingTime(t *testing.T) {
	handler, err := NewTimeMachineHandler(&timeMachineReaderStub{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTimeMachineHandler() error = %v", err)
	}
	router := timeMachineRouter(handler)

	request := httptest.NewRequest(http.MethodGet, "/v1/time-machine/hosts/node-01", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestTimeMachineHandlerReturnsServiceUnavailable(t *testing.T) {
	handler, err := NewTimeMachineHandler(&timeMachineReaderStub{err: errors.New("database unavailable")}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTimeMachineHandler() error = %v", err)
	}
	router := timeMachineRouter(handler)

	request := httptest.NewRequest(http.MethodGet, "/v1/time-machine/hosts/node-01?at=2026-07-15T12:00:00Z", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func timeMachineRouter(handler *TimeMachineHandler) http.Handler {
	router := chi.NewRouter()
	router.Get("/v1/time-machine/hosts/{hostname}", handler.Snapshot)
	return router
}
