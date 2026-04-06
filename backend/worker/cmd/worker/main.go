package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/chetasparekh/gotube-lite/api/pkg/queue"
	"github.com/chetasparekh/gotube-lite/api/pkg/repository"
	"github.com/chetasparekh/gotube-lite/api/pkg/storage"
	"github.com/chetasparekh/gotube-lite/worker/internal/config"
	"github.com/chetasparekh/gotube-lite/worker/internal/consumer"
	"github.com/chetasparekh/gotube-lite/worker/internal/processor"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create temp dir")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	dbpool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer dbpool.Close()
	log.Info().Msg("connected to PostgreSQL")

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr(),
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer rdb.Close()
	log.Info().Msg("connected to Redis")

	// Storage
	store, err := storage.NewS3Storage(storage.Config{
		Endpoint:       cfg.Storage.Endpoint,
		PublicEndpoint: cfg.Storage.PublicEndpoint,
		AccessKey:      cfg.Storage.AccessKey,
		SecretKey:      cfg.Storage.SecretKey,
		UsePathStyle:   cfg.Storage.UsePathStyle,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize object storage")
	}

	// Queue
	q := queue.NewQueue(rdb)
	if err := q.CreateConsumerGroup(ctx); err != nil {
		log.Warn().Err(err).Msg("consumer group creation (may already exist)")
	}

	// Repositories
	videoRepo := repository.NewVideoRepository(dbpool)
	jobRepo := repository.NewJobRepository(dbpool)

	// Processor
	proc := processor.NewVideoProcessor(videoRepo, jobRepo, store, cfg)

	// Consumer
	cons := consumer.NewConsumer(q, proc, cfg.Concurrency)

	// Metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		log.Info().Msg("worker metrics server on :9091")
		if err := http.ListenAndServe(":9091", mux); err != nil {
			log.Error().Err(err).Msg("metrics server failed")
		}
	}()

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("received shutdown signal")
		cancel()
	}()

	log.Info().Int("concurrency", cfg.Concurrency).Msg("starting worker")
	if err := cons.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("consumer failed")
	}
	log.Info().Msg("worker stopped")
}
