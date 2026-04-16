package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/handler"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/messaging"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/repository/postgres"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/repository/redis"
	"github.com/leonardo-gorska/nexuslink/internal/app"
	"github.com/leonardo-gorska/nexuslink/pkg/config"
	"github.com/leonardo-gorska/nexuslink/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redisclient "github.com/redis/go-redis/v9"
)

func main() {
	// 1. Load config
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	// 2. Initialize logger
	logger.New(cfg.AppEnv, cfg.LogLevel)
	slog.Info("starting nexuslink API...", slog.String("env", cfg.AppEnv))

	parentCtx := context.Background()

	// 3. Connect to Database
	dbPool, err := pgxpool.New(parentCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(parentCtx); err != nil {
		slog.Error("database ping failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("database connection established")

	// 3.5. Connect to Redis
	redisOpts, err := redisclient.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse redis url", slog.Any("error", err))
		os.Exit(1)
	}
	redisCli := redisclient.NewClient(redisOpts)
	defer redisCli.Close()

	if err := redisCli.Ping(parentCtx).Err(); err != nil {
		slog.Error("redis ping failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("redis connection established")

	cacheStore := redis.NewCacheStore(redisCli)
	rateLimiterStore := redis.NewRateLimiter(redisCli)

	// Messaging Setup
	rmqConn, err := messaging.NewConnection(cfg.RabbitMQURL)
	var publisher *messaging.Publisher
	if err != nil {
		slog.Warn("rabbitmq not available, running without publisher", slog.Any("error", err))
	} else {
		defer rmqConn.Close()
		messaging.SetupInfrastructure(rmqConn.Conn)
		publisher, err = messaging.NewPublisher(rmqConn)
		if err != nil {
			slog.Warn("rabbitmq publisher not created", slog.Any("error", err))
			publisher = nil
		} else {
			slog.Info("rabbitmq connection established and publisher ready")
		}
	}

	// 4. Setup dependency injection
	linkRepo := postgres.NewLinkRepository(dbPool)
	linkService := app.NewLinkService(linkRepo, cacheStore, cfg.HashLength)

	// Handlers
	linkHandler := handler.NewLinkHandler(linkService)
	// Passing publisher into RedirectHandler
	redirectHandler := handler.NewRedirectHandler(linkService, publisher)
	healthHandler := handler.NewHealthHandler()
	analyticsHandler := handler.NewAnalyticsHandler(nil) // Phase 3

	// Router
	router := handler.NewRouter(linkHandler, redirectHandler, healthHandler, analyticsHandler, rateLimiterStore)

	// 5. Setup HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Prometheus Metrics Server (Port 9090)
	metricsSrv := &http.Server{
		Addr:    ":" + cfg.MetricsPort,
		Handler: promhttp.Handler(),
	}

	go func() {
		slog.Info("metrics server listening", slog.String("port", cfg.MetricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server failed", slog.Any("error", err))
		}
	}()

	// 6. Graceful Shutdown
	go func() {
		slog.Info("server listening", slog.String("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	if err := metricsSrv.Shutdown(ctx); err != nil {
		slog.Warn("metrics server shutdown with error", slog.Any("error", err))
	}

	slog.Info("server exited")
}
