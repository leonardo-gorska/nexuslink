package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/messaging"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/repository/postgres"
	"github.com/leonardo-gorska/nexuslink/internal/app"
	"github.com/leonardo-gorska/nexuslink/pkg/config"
	"github.com/leonardo-gorska/nexuslink/pkg/logger"
	"github.com/leonardo-gorska/nexuslink/pkg/metrics"
	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	logger.New(cfg.AppEnv, cfg.LogLevel)
	slog.Info("starting nexuslink analytics worker...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Database
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		slog.Error("database ping failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("postgres connected")

	// MaxMind GeoIP
	var geoReader *geoip2.Reader
	geoPath := os.Getenv("GEOIP_DB_PATH")
	if geoPath != "" {
		geoReader, err = geoip2.Open(geoPath)
		if err != nil {
			slog.Warn("could not load GeoIP DB, using nil", slog.String("path", geoPath), slog.Any("error", err))
		} else {
			defer geoReader.Close()
		}
	}

	// Repo & Service
	analyticsRepo := postgres.NewAnalyticsRepository(dbPool)
	analyticsService := app.NewAnalyticsService(analyticsRepo, geoReader)

	// Messaging Setup
	rmqConn, err := messaging.NewConnection(cfg.RabbitMQURL)
	if err != nil {
		slog.Error("failed to connect to rabbitmq", slog.Any("error", err))
		os.Exit(1)
	}
	defer rmqConn.Close()

	if err := messaging.SetupInfrastructure(rmqConn.Conn); err != nil {
		slog.Error("failed to setup rabbitmq topology", slog.Any("error", err))
		os.Exit(1)
	}

	consumer, err := messaging.NewConsumer(rmqConn)
	if err != nil {
		slog.Error("failed to create consumer", slog.Any("error", err))
		os.Exit(1)
	}

	batchSize, _ := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	if batchSize == 0 {
		batchSize = 500
	}
	
	// Timeout must be provided, eg BATCH_TIMEOUT string parsing.
	// We'll hardcode to 5s if not specfied.
	batchTimeout := 5 * time.Second
	bt, _ := time.ParseDuration(os.Getenv("BATCH_TIMEOUT"))
	if bt > 0 {
		batchTimeout = bt
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Metrics Server
	metricsSrv := &http.Server{
		Addr:    ":9091",
		Handler: promhttp.Handler(),
	}

	go func() {
		slog.Info("metrics server listening", slog.String("port", "9091"))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server error", slog.Any("error", err))
		}
	}()

	// Worker Loop
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				batch, err := consumer.Consume(ctx, batchSize, batchTimeout)
				if err != nil {
					slog.Error("consumer closed or errored", slog.Any("error", err))
					return
				}
				if len(batch) > 0 {
					slog.Info("processing batch", slog.Int("size", len(batch)))
					err = analyticsService.ProcessClickBatch(ctx, batch)
					if err != nil {
						metrics.EventsProcessedTotal.WithLabelValues("error").Add(float64(len(batch)))
						slog.Error("failed to process batch", slog.Any("error", err))
						consumer.Nack(uint64(batch[len(batch)-1].ID), false)
					} else {
						metrics.EventsProcessedTotal.WithLabelValues("success").Add(float64(len(batch)))
						if ackErr := consumer.Ack(uint64(batch[len(batch)-1].ID)); ackErr != nil {
							slog.Error("failed to ack batch", slog.Any("error", ackErr))
						} else {
							slog.Info("successfully acked batch")
						}
					}
				}
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down worker gracefully, waiting for current batch to finish...")

	cancel() 
	
	metricsSrv.Shutdown(context.Background())
	wg.Wait()
	slog.Info("worker exited")
}
