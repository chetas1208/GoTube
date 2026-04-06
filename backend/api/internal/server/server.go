package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/chetasparekh/gotube-lite/api/internal/config"
	"github.com/chetasparekh/gotube-lite/api/internal/handler"
	mw "github.com/chetasparekh/gotube-lite/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(
	cfg *config.Config,
	authH *handler.AuthHandler,
	videoH *handler.VideoHandler,
	commentH *handler.CommentHandler,
	healthH *handler.HealthHandler,
) http.Handler {
	r := chi.NewRouter()
	allowedOrigins := []string{"http://localhost:3000"}
	if cfg.API.CORSAllowedOrigins != "" {
		allowedOrigins = strings.Split(cfg.API.CORSAllowedOrigins, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mw.Logging)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// Health
	r.Get("/health", healthH.Health)
	r.Get("/ready", healthH.Ready)

	authRequired := mw.AuthRequired(cfg.JWT.Secret)
	optionalAuth := mw.OptionalAuth(cfg.JWT.Secret)

	authRateLimit := httprate.LimitByIP(cfg.RateLimit.AuthRPM, time.Minute)
	uploadRateLimit := httprate.LimitByIP(cfg.RateLimit.UploadRPM, time.Minute)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.With(authRateLimit).Post("/register", authH.Register)
			r.With(authRateLimit).Post("/login", authH.Login)
			r.With(authRequired).Post("/logout", authH.Logout)
			r.With(authRequired).Get("/me", authH.Me)
			r.Post("/refresh", authH.Refresh)
		})

		// Video routes
		r.Route("/videos", func(r chi.Router) {
			r.With(optionalAuth).Get("/", videoH.ListRecent)

			r.With(authRequired, uploadRateLimit).Post("/initiate-upload", videoH.InitiateUpload)

			r.With(authRequired).Get("/my", videoH.ListMyVideos)

			r.Route("/{id}", func(r chi.Router) {
				r.With(optionalAuth).Get("/", videoH.Get)
				r.With(authRequired).Post("/complete-upload", videoH.CompleteUpload)
				r.With(authRequired).Patch("/", videoH.Update)
				r.With(authRequired).Delete("/", videoH.Delete)

				r.With(optionalAuth).Post("/view", videoH.RecordView)
				r.With(authRequired).Post("/like", videoH.Like)
				r.With(authRequired).Delete("/like", videoH.Unlike)

				r.With(optionalAuth).Get("/playback", videoH.Playback)

				// Comments
				r.With(optionalAuth).Get("/comments", commentH.List)
				r.With(authRequired).Post("/comments", commentH.Create)
			})
		})

		// Search & Trending
		r.With(optionalAuth).Get("/search", videoH.Search)
		r.With(optionalAuth).Get("/trending", videoH.Trending)
	})

	return r
}
