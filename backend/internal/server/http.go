package server

import (
	"net/http"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/api"
	"go.uber.org/zap"
)

// buildHTTPServer constructs the HTTP server with the assembled router and timeouts.
func buildHTTPServer(deps *Dependencies, logger *zap.Logger, addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(api.Handlers{Health: deps.Health, Metrics: deps.Metrics, Events: deps.Events, Dashboard: deps.Dashboard, Replay: deps.Replay, TimeMachine: deps.TimeMachine, AI: deps.AI}, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
