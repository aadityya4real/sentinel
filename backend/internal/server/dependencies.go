// Package server bootstraps the Sentinel application and wires its HTTP server.
package server

import (
	"fmt"

	"github.com/aadityya4real/sentinel/backend/internal/ai"
	"github.com/aadityya4real/sentinel/backend/internal/api"
	"github.com/aadityya4real/sentinel/backend/internal/collector"
	"github.com/aadityya4real/sentinel/backend/internal/config"
	"github.com/aadityya4real/sentinel/backend/internal/dashboard"
	"github.com/aadityya4real/sentinel/backend/internal/database"
	eventing "github.com/aadityya4real/sentinel/backend/internal/events"
	"github.com/aadityya4real/sentinel/backend/internal/replay"
	"github.com/aadityya4real/sentinel/backend/internal/redis"
	"github.com/aadityya4real/sentinel/backend/internal/storage"
	"github.com/aadityya4real/sentinel/backend/internal/timemachine"
	"go.uber.org/zap"
)

// Dependencies holds all HTTP handlers built from the service graph.
type Dependencies struct {
	Health      *api.HealthHandler
	Metrics     *api.MetricsHandler
	Events      *api.EventsHandler
	Dashboard   *api.DashboardHandler
	Replay      *api.ReplayHandler
	TimeMachine *api.TimeMachineHandler
	AI          *api.AIHandler
}

// buildDependencies wires the full dependency graph and returns assembled HTTP handlers.
func buildDependencies(cfg *config.Config, db *database.Database, redisClient *redis.Redis, log *zap.Logger) (*Dependencies, error) {
	repository, err := storage.NewPostgreSQLMetricsRepository(db.Pool)
	if err != nil {
		return nil, fmt.Errorf("create metrics repository: %w", err)
	}
	cache, err := storage.NewRedisLatestMetricsCache(redisClient.Client)
	if err != nil {
		return nil, fmt.Errorf("create metrics cache: %w", err)
	}
	latestEventCache, err := storage.NewRedisLatestEventCache(redisClient.Client)
	if err != nil {
		return nil, fmt.Errorf("create events cache: %w", err)
	}
	events, err := storage.NewPostgreSQLEventStore(db.Pool)
	if err != nil {
		return nil, fmt.Errorf("create event store: %w", err)
	}

	eventCollector, err := eventing.NewCollector(events, latestEventCache)
	if err != nil {
		return nil, fmt.Errorf("create event collector: %w", err)
	}
	replayEngine, err := replay.NewEngine(events)
	if err != nil {
		return nil, fmt.Errorf("create replay engine: %w", err)
	}
	timeMachineEngine, err := timemachine.NewEngine(events)
	if err != nil {
		return nil, fmt.Errorf("create time machine engine: %w", err)
	}

	analyzer, err := buildAnalyzer(cfg, events)
	if err != nil {
		return nil, fmt.Errorf("create AI analyzer: %w", err)
	}

	service, err := collector.NewService(repository, events, cache)
	if err != nil {
		return nil, fmt.Errorf("create collector service: %w", err)
	}

	dashboardRepository, err := storage.NewPostgreSQLDashboardRepository(db.Pool)
	if err != nil {
		return nil, fmt.Errorf("create dashboard repository: %w", err)
	}
	dashboardService, err := dashboard.NewService(dashboardRepository)
	if err != nil {
		return nil, fmt.Errorf("create dashboard service: %w", err)
	}

	healthHandler, err := api.NewHealthHandler(db, redisClient, log)
	if err != nil {
		return nil, fmt.Errorf("create health handler: %w", err)
	}
	metricsHandler, err := api.NewMetricsHandler(service, log)
	if err != nil {
		return nil, fmt.Errorf("create metrics handler: %w", err)
	}
	eventsHandler, err := api.NewEventsHandler(eventCollector, log)
	if err != nil {
		return nil, fmt.Errorf("create events handler: %w", err)
	}
	dashboardHandler, err := api.NewDashboardHandler(dashboardService, log)
	if err != nil {
		return nil, fmt.Errorf("create dashboard handler: %w", err)
	}
	replayHandler, err := api.NewReplayHandler(replayEngine, log)
	if err != nil {
		return nil, fmt.Errorf("create replay handler: %w", err)
	}
	timeMachineHandler, err := api.NewTimeMachineHandler(timeMachineEngine, log)
	if err != nil {
		return nil, fmt.Errorf("create time machine handler: %w", err)
	}
	aiHandler, err := api.NewAIHandler(analyzer, log)
	if err != nil {
		return nil, fmt.Errorf("create AI handler: %w", err)
	}

	return &Dependencies{
		Health:      healthHandler,
		Metrics:     metricsHandler,
		Events:      eventsHandler,
		Dashboard:   dashboardHandler,
		Replay:      replayHandler,
		TimeMachine: timeMachineHandler,
		AI:          aiHandler,
	}, nil
}

// buildAnalyzer returns the configured AI analyzer, or a disabled stub when AI is off.
func buildAnalyzer(cfg *config.Config, events *storage.PostgreSQLEventStore) (ai.Analyzer, error) {
	if !cfg.AIEnabled {
		return ai.DisabledAnalyzer{}, nil
	}
	client, err := ai.NewOpenAICompatibleClient(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel)
	if err != nil {
		return nil, err
	}
	return ai.NewService(events, client)
}
