package api

import (
	"encoding/json"
	"net/http"

	"github.com/aadityya4real/sentinel/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Handlers bundles all HTTP handlers required by the Sentinel API router.
type Handlers struct {
	Health       *HealthHandler
	Metrics      *MetricsHandler
	Events       *EventsHandler
	Dashboard    *DashboardHandler
	Replay       *ReplayHandler
	TimeMachine  *TimeMachineHandler
	AI           *AIHandler
}

// NewRouter creates the Sentinel HTTP router with standardized /api/v1 routes.
func NewRouter(handlers Handlers, logger *zap.Logger) http.Handler {
	r := chi.NewRouter()
	for _, mw := range middleware.Chain(logger) {
		r.Use(mw)
	}

	r.Get("/", landingHandler(apiVersion))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.Health.ServeHTTP)
		r.Post("/metrics", handlers.Metrics.ServeHTTP)
		r.Post("/events", handlers.Events.ServeHTTP)

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/overview", handlers.Dashboard.Overview)
			r.Get("/hosts", handlers.Dashboard.Hosts)
			r.Get("/hosts/{hostname}/metrics", handlers.Dashboard.History)
		})

		r.Route("/replay", func(r chi.Router) {
			r.Get("/hosts/{hostname}", handlers.Replay.Replay)
		})

		r.Route("/time-machine", func(r chi.Router) {
			r.Get("/hosts/{hostname}", handlers.TimeMachine.Snapshot)
		})

		r.Route("/ai", func(r chi.Router) {
			r.Post("/incidents/analyze", handlers.AI.AnalyzeIncident)
		})
	})

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
