package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/solid-vpn/api/config"
	"github.com/solid-vpn/api/internal/health"
	"github.com/solid-vpn/api/routes"
)

func main() {
	// Load .env if present (non-fatal in production where env vars are injected)
	_ = godotenv.Load()

	// Bootstrap a temporary logger for startup errors
	startupLog, _ := zap.NewProduction()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		startupLog.Fatal("failed to load configuration", zap.Error(err))
	}

	// Build the structured logger
	log, err := buildLogger(cfg.LogLevel)
	if err != nil {
		startupLog.Fatal("failed to build logger", zap.Error(err))
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting vpn-api",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
	)

	// Connect to PostgreSQL
	pool, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		// Non-fatal during development — the server still starts but /ready will report unhealthy.
		log.Warn("database connection failed; /ready will report not-ready", zap.Error(err))
	}
	if pool != nil {
		defer pool.Close()
	}

	// Build handlers
	healthHandler := health.NewHandler(pool)

	// Build router
	handler := routes.New(routes.Options{
		Logger:        log,
		HealthHandler: healthHandler,
	})

	// HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info("vpn-api listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Graceful shutdown on signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-serverErrors:
		log.Fatal("server error", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	} else {
		log.Info("vpn-api stopped cleanly")
	}
}

// connectDB establishes a pgxpool connection and verifies it with a ping.
func connectDB(databaseURL string, log *zap.Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}

	log.Info("database connected")
	return pool, nil
}

// buildLogger creates a zap.Logger at the requested level.
func buildLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoderCfg,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapCfg.Build()
}
