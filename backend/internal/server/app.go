package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/config"
	"github.com/aadityya4real/sentinel/backend/internal/database"
	"github.com/aadityya4real/sentinel/backend/internal/logger"
	"github.com/aadityya4real/sentinel/backend/internal/redis"
	"go.uber.org/zap"
)

const startupTimeout = 15 * time.Second

// App owns the long-lived dependencies and HTTP server for the Sentinel process.
type App struct {
	cfg    *config.Config
	log    *zap.Logger
	db     *database.Database
	redis  *redis.Redis
	deps   *Dependencies
}

// New validates configuration and builds the application and all of its dependencies.
func New(cfg *config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log, err := logger.New()
	if err != nil {
		return nil, err
	}

	app := &App{cfg: cfg, log: log}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), startupTimeout)
	defer cancelStartup()

	db, err := database.New(startupCtx, cfg.DatabaseURL)
	if err != nil {
		app.cleanup()
		return nil, err
	}
	app.db = db

	if err := db.ApplyMigrations(startupCtx); err != nil {
		app.cleanup()
		return nil, err
	}

	redisClient, err := redis.New(startupCtx, cfg.RedisAddress, cfg.RedisPassword)
	if err != nil {
		app.cleanup()
		return nil, err
	}
	app.redis = redisClient

	deps, err := buildDependencies(cfg, db, redisClient, log)
	if err != nil {
		app.cleanup()
		return nil, err
	}
	app.deps = deps

	return app, nil
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
func (a *App) Run() error {
	addr := ":" + a.cfg.Port
	httpServer := buildHTTPServer(a.deps, a.log, addr)

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		a.log.Info("Sentinel server starting", zap.String("address", addr))
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			a.cleanup()
			return err
		}
	case <-shutdownCtx.Done():
		a.log.Info("shutdown signal received")
	}

	gracefulCtx, cancelGraceful := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelGraceful()
	if err := httpServer.Shutdown(gracefulCtx); err != nil {
		a.log.Error("graceful HTTP shutdown", zap.Error(err))
	}

	a.cleanup()
	return nil
}

// cleanup releases all owned infrastructure connections. Safe to call multiple times.
func (a *App) cleanup() {
	if a.deps != nil && a.deps.Hub != nil {
		a.deps.Hub.Close()
	}
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			a.log.Error("close Redis", zap.Error(err))
		}
		a.redis = nil
	}
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}
	_ = a.log.Sync()
}
