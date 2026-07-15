// Command server runs the Sentinel Collector HTTP service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/ai"
	"github.com/aadityya4real/sentinel/backend/internal/api"
	"github.com/aadityya4real/sentinel/backend/internal/collector"
	"github.com/aadityya4real/sentinel/backend/internal/config"
	"github.com/aadityya4real/sentinel/backend/internal/dashboard"
	"github.com/aadityya4real/sentinel/backend/internal/database"
	eventing "github.com/aadityya4real/sentinel/backend/internal/events"
	"github.com/aadityya4real/sentinel/backend/internal/logger"
	"github.com/aadityya4real/sentinel/backend/internal/redis"
	"github.com/aadityya4real/sentinel/backend/internal/replay"
	"github.com/aadityya4real/sentinel/backend/internal/storage"
	"github.com/aadityya4real/sentinel/backend/internal/timemachine"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logg, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}
	defer logg.Sync()

	startupContext, cancelStartup := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStartup()

	db, err := database.New(startupContext, cfg.DatabaseURL)
	if err != nil {
		logg.Fatal("connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()
	if err := db.ApplyMigrations(startupContext); err != nil {
		logg.Fatal("apply PostgreSQL migrations", zap.Error(err))
	}

	redisClient, err := redis.New(startupContext, cfg.RedisAddress, cfg.RedisPassword)
	if err != nil {
		logg.Fatal("connect to Redis", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logg.Error("close Redis", zap.Error(err))
		}
	}()

	repository, err := storage.NewPostgreSQLMetricsRepository(db.Pool)
	if err != nil {
		logg.Fatal("create PostgreSQL metrics repository", zap.Error(err))
	}
	cache, err := storage.NewRedisLatestMetricsCache(redisClient.Client)
	if err != nil {
		logg.Fatal("create Redis metrics cache", zap.Error(err))
	}
	latestEventCache, err := storage.NewRedisLatestEventCache(redisClient.Client)
	if err != nil {
		logg.Fatal("create Redis latest event cache", zap.Error(err))
	}
	events, err := storage.NewPostgreSQLEventStore(db.Pool)
	if err != nil {
		logg.Fatal("create PostgreSQL event store", zap.Error(err))
	}
	eventCollector, err := eventing.NewCollector(events, latestEventCache)
	if err != nil {
		logg.Fatal("create event collector", zap.Error(err))
	}
	replayEngine, err := replay.NewEngine(events)
	if err != nil {
		logg.Fatal("create replay engine", zap.Error(err))
	}
	timeMachineEngine, err := timemachine.NewEngine(events)
	if err != nil {
		logg.Fatal("create time machine engine", zap.Error(err))
	}
	var analyzer ai.Analyzer = ai.DisabledAnalyzer{}
	if cfg.AIEnabled {
		client, err := ai.NewOpenAICompatibleClient(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel)
		if err != nil {
			logg.Fatal("create AI client", zap.Error(err))
		}
		analyzer, err = ai.NewService(events, client)
		if err != nil {
			logg.Fatal("create AI incident analyzer", zap.Error(err))
		}
	}
	service, err := collector.NewService(repository, events, cache)
	if err != nil {
		logg.Fatal("create collector service", zap.Error(err))
	}
	metricsHandler, err := api.NewMetricsHandler(service, logg)
	if err != nil {
		logg.Fatal("create metrics handler", zap.Error(err))
	}
	eventsHandler, err := api.NewEventsHandler(eventCollector, logg)
	if err != nil {
		logg.Fatal("create events handler", zap.Error(err))
	}
	dashboardRepository, err := storage.NewPostgreSQLDashboardRepository(db.Pool)
	if err != nil {
		logg.Fatal("create PostgreSQL dashboard repository", zap.Error(err))
	}
	dashboardService, err := dashboard.NewService(dashboardRepository)
	if err != nil {
		logg.Fatal("create dashboard service", zap.Error(err))
	}
	dashboardHandler, err := api.NewDashboardHandler(dashboardService, logg)
	if err != nil {
		logg.Fatal("create dashboard handler", zap.Error(err))
	}
	replayHandler, err := api.NewReplayHandler(replayEngine, logg)
	if err != nil {
		logg.Fatal("create replay handler", zap.Error(err))
	}
	timeMachineHandler, err := api.NewTimeMachineHandler(timeMachineEngine, logg)
	if err != nil {
		logg.Fatal("create time machine handler", zap.Error(err))
	}
	aiHandler, err := api.NewAIHandler(analyzer, logg)
	if err != nil {
		logg.Fatal("create AI handler", zap.Error(err))
	}
	healthHandler, err := api.NewHealthHandler(cache, logg)
	if err != nil {
		logg.Fatal("create health handler", zap.Error(err))
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:              addr,
			Handler:           api.NewRouter(metricsHandler, eventsHandler, dashboardHandler, replayHandler, timeMachineHandler, aiHandler, healthHandler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logg.Info("Sentinel server starting", zap.String("address", addr))
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logg.Fatal("serve HTTP", zap.Error(err))
		}
	case <-shutdownContext.Done():
		logg.Info("shutdown signal received")
	}

	gracefulShutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(gracefulShutdownContext); err != nil {
		logg.Error("graceful HTTP shutdown", zap.Error(err))
	}
}
