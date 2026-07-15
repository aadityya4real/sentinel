package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates the Sentinel HTTP router and registers its API endpoints.
func NewRouter(metricsHandler *MetricsHandler, eventsHandler *EventsHandler, dashboardHandler *DashboardHandler, replayHandler *ReplayHandler, timeMachineHandler *TimeMachineHandler, aiHandler *AIHandler, healthHandler *HealthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Get("/health", healthHandler.ServeHTTP)
	r.Post("/v1/metrics", metricsHandler.ServeHTTP)
	r.Post("/api/v1/events", eventsHandler.ServeHTTP)
	r.Get("/v1/dashboard/overview", dashboardHandler.Overview)
	r.Get("/v1/dashboard/hosts", dashboardHandler.Hosts)
	r.Get("/v1/dashboard/hosts/{hostname}/metrics", dashboardHandler.History)
	r.Get("/v1/replay/hosts/{hostname}", replayHandler.Replay)
	r.Get("/v1/time-machine/hosts/{hostname}", timeMachineHandler.Snapshot)
	r.Post("/v1/ai/incidents/analyze", aiHandler.AnalyzeIncident)

	return r
}

func writeJSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
