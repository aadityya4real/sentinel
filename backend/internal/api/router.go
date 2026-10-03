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
	Health      *HealthHandler
	Metrics     *MetricsHandler
	Events      *EventsHandler
	Dashboard   *DashboardHandler
	Replay      *ReplayHandler
	TimeMachine *TimeMachineHandler
	AI          *AIHandler
	Websocket   *WebsocketHandler
}

// RouterSecurity controls access to protected operations and browser origins.
type RouterSecurity struct {
	APIToken       string
	AllowedOrigins []string
}

// NewRouter creates the Sentinel HTTP router with standardized /api/v1 routes.
func NewRouter(handlers Handlers, logger *zap.Logger, security RouterSecurity) http.Handler {
	r := chi.NewRouter()
	for _, mw := range middleware.Chain(logger, security.AllowedOrigins) {
		r.Use(mw)
	}

	r.Get("/", landingHandler(apiVersion))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.Health.ServeHTTP)

		r.Group(func(protected chi.Router) {
			protected.Use(middleware.RequireBearerToken(security.APIToken))
			protected.Post("/metrics", handlers.Metrics.ServeHTTP)
			protected.Post("/events", handlers.Events.ServeHTTP)
			protected.Post("/ai/incidents/analyze", handlers.AI.AnalyzeIncident)
		})
		r.Get("/events", handlers.Events.List)

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

	})

	r.Get("/ws/v1/metrics", handlers.Websocket.Metrics)

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
