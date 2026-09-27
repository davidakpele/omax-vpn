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
	"github.com/solid-vpn/api/internal/auth"
	"github.com/solid-vpn/api/internal/devices"
	"github.com/solid-vpn/api/internal/engine"
	"github.com/solid-vpn/api/internal/health"
	"github.com/solid-vpn/api/internal/servers"
	"github.com/solid-vpn/api/internal/users"
	"github.com/solid-vpn/api/internal/vpn"
	"github.com/solid-vpn/api/routes"
)

func main() {
	_ = godotenv.Load()

	startupLog, _ := zap.NewProduction()

	cfg, err := config.Load()
	if err != nil {
		startupLog.Fatal("failed to load configuration", zap.Error(err))
	}

	log, err := buildLogger(cfg.LogLevel)
	if err != nil {
		startupLog.Fatal("failed to build logger", zap.Error(err))
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting vpn-api",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
	)

	pool, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Warn("database connection failed; /ready will report not-ready", zap.Error(err))
	}
	if pool != nil {
		defer pool.Close()
	}

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, auth.ServiceConfig{
		JWTSecret:          cfg.JWTSecret,
		AccessExpiryMin:    cfg.JWTAccessExpiryMin,
		RefreshExpiryHours: cfg.JWTRefreshExpiryH,
	})
	authHandler := auth.NewHandler(authSvc)

	userRepo := users.NewRepository(pool)
	userSvc := users.NewService(userRepo)
	userHandler := users.NewHandler(userSvc)

	deviceRepo := devices.NewRepository(pool)
	deviceSvc := devices.NewService(deviceRepo)
	deviceHandler := devices.NewHandler(deviceSvc)

	serverRepo := servers.NewRepository(pool)
	serverSvc := servers.NewService(serverRepo)
	serverHandler := servers.NewHandler(serverSvc)

	engineClient := engine.NewClient(cfg.VPNEngineURL, cfg.VPNEngineToken)

	vpnRepo := vpn.NewRepository(pool)
	vpnSvc := vpn.NewService(vpnRepo, serverSvc, deviceRepo, engineClient, cfg.VPNDNS)
	vpnHandler := vpn.NewHandler(vpnSvc)

	handler := routes.New(routes.Options{
		Logger:        log,
		HealthHandler: health.NewHandler(pool, engineClient),
		AuthHandler:   authHandler,
		AuthService:   authSvc,
		UserHandler:   userHandler,
		DeviceHandler: deviceHandler,
		ServerHandler: serverHandler,
		VPNHandler:    vpnHandler,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("vpn-api listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

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
