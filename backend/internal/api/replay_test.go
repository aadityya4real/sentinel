package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aadityya4real/sentinel/backend/internal/replay"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type replayReaderStub struct {
	timeline replay.Timeline
	err      error
}

func (s *replayReaderStub) Replay(context.Context, replay.Request) (replay.Timeline, error) {
	return s.timeline, s.err
}

func TestReplayHandlerReturnsTimeline(t *testing.T) {
	handler, err := NewReplayHandler(&replayReaderStub{timeline: replay.Timeline{Hostname: "node-01"}}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewReplayHandler() error = %v", err)
	}
	router := chi.NewRouter()
	router.Get("/v1/replay/hosts/{hostname}", handler.Replay)

	request := httptest.NewRequest(http.MethodGet, "/v1/replay/hosts/node-01", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestReplayHandlerReturnsBadRequestForInvalidLimit(t *testing.T) {
	handler, err := NewReplayHandler(&replayReaderStub{}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewReplayHandler() error = %v", err)
	}
	router := chi.NewRouter()
	router.Get("/v1/replay/hosts/{hostname}", handler.Replay)

	request := httptest.NewRequest(http.MethodGet, "/v1/replay/hosts/node-01?limit=1000", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestReplayHandlerReturnsServiceUnavailable(t *testing.T) {
	handler, err := NewReplayHandler(&replayReaderStub{err: errors.New("database unavailable")}, zap.NewNop())
	if err != nil {
		t.Fatalf("NewReplayHandler() error = %v", err)
	}
	router := chi.NewRouter()
	router.Get("/v1/replay/hosts/{hostname}", handler.Replay)

	request := httptest.NewRequest(http.MethodGet, "/v1/replay/hosts/node-01", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
