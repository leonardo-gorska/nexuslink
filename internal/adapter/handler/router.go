package handler

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/handler/middleware"
)

func NewRouter(
	linkHandler *LinkHandler,
	redirectHandler *RedirectHandler,
	healthHandler *HealthHandler,
	analyticsHandler *AnalyticsHandler,
	rateLimiter middleware.RateLimiterStore,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS)
	r.Use(middleware.Metrics)

	// Health and Readiness probes
	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	// Redirect path - Hot Path
	r.With(middleware.RateLimiter(rateLimiter, 1000, time.Minute)).Get("/r/{hash}", redirectHandler.Redirect)

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Links
		r.Route("/links", func(r chi.Router) {
			r.With(middleware.RateLimiter(rateLimiter, 10, time.Minute)).Post("/", linkHandler.CreateLink)
			r.With(middleware.RateLimiter(rateLimiter, 60, time.Minute)).Get("/{hash}", linkHandler.GetLink)
			r.With(middleware.RateLimiter(rateLimiter, 10, time.Minute)).Delete("/{hash}", linkHandler.DeleteLink)
			
			// Analytics
			r.With(middleware.RateLimiter(rateLimiter, 30, time.Minute)).Get("/{hash}/stats", analyticsHandler.GetAnalytics)
		})
	})

	return r
}
