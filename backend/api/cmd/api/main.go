package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chetasparekh/gotube-lite/api/internal/config"
	"github.com/chetasparekh/gotube-lite/api/internal/handler"
	"github.com/chetasparekh/gotube-lite/api/internal/repository"
	"github.com/chetasparekh/gotube-lite/api/internal/server"
	"github.com/chetasparekh/gotube-lite/api/internal/service"
	"github.com/chetasparekh/gotube-lite/api/pkg/queue"
	"github.com/chetasparekh/gotube-lite/api/pkg/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Structured logging setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("API_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Handle migrations subcommand
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations(cfg)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		runSeed(cfg)
		return
	}

	ctx := context.Background()

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

	// Object storage
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
		log.Warn().Err(err).Msg("failed to create consumer group (may already exist)")
	}

	// Repositories
	userRepo := repository.NewUserRepository(dbpool)
	videoRepo := repository.NewVideoRepository(dbpool)
	commentRepo := repository.NewCommentRepository(dbpool)
	likeRepo := repository.NewLikeRepository(dbpool)
	viewRepo := repository.NewViewRepository(dbpool)
	jobRepo := repository.NewJobRepository(dbpool)
	uploadRepo := repository.NewUploadSessionRepository(dbpool)
	refreshRepo := repository.NewRefreshTokenRepository(dbpool)

	// Services
	authSvc := service.NewAuthService(userRepo, refreshRepo, cfg.JWT)
	videoSvc := service.NewVideoService(videoRepo, uploadRepo, jobRepo, likeRepo, viewRepo, store, q, cfg.Storage, cfg.Upload)
	commentSvc := service.NewCommentService(commentRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, cfg.JWT.RefreshTTL)
	videoHandler := handler.NewVideoHandler(videoSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	healthHandler := handler.NewHealthHandler(dbpool, rdb)

	// Router
	router := server.NewRouter(cfg, authHandler, videoHandler, commentHandler, healthHandler)

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.API.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("server shutdown error")
		}
	}()

	log.Info().Str("port", cfg.API.Port).Msg("starting GoTube API server")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server failed")
	}
}

func runMigrations(cfg *config.Config) {
	direction := "up"
	if len(os.Args) > 2 {
		direction = os.Args[2]
	}
	fmt.Printf("Running migrations %s on %s\n", direction, cfg.DB.DSN())
	// In production, use golang-migrate/migrate CLI or integrate it here
	fmt.Println("Use: migrate -path ./migrations -database", cfg.DB.DSN(), "-verbose", direction)
}

func runSeed(cfg *config.Config) {
	fmt.Println("Seeding database at", cfg.DB.DSN())
	fmt.Println("Seed functionality not yet implemented. Add seed data via SQL scripts.")
}
