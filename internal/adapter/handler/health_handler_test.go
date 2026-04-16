package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/handler"
	"github.com/redis/go-redis/v9"
)

func TestHealthz(t *testing.T) {
	h := handler.NewHealthHandler(nil, nil, nil)
	router := chi.NewRouter()
	router.Get("/healthz", h.Healthz)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}
}

func TestReadyz(t *testing.T) {
	// Because of the dependencies on *pgxpool.Pool and *redis.Client matching underlying actual struct logic
	// in pinging, unit testing this specific readiness with nil pointers usually causes panics in the real implementation
	// if we don't mock the specific calls, but pure unit test of it requires actual interfaces.
	// Typically readiness probe depends on actual connection objects. 
	// We will just verify it serves 503 if we send nil interfaces which triggers panic recovery usually or explicit errors.
	
	// Assuming the implementation handles nil gracefully and returns 503:
	cfg, err := pgxpool.ParseConfig("postgres://invalid")
	if err != nil {
		t.Skip("Parsing fail")
	}
	db, _ := pgxpool.NewWithConfig(context.Background(), cfg)
	
	rdb := redis.NewClient(&redis.Options{})

	h := handler.NewHealthHandler(db, rdb, nil)
	router := chi.NewRouter()
	router.Get("/readyz", h.Readyz)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d due to invalid connections; got %d", http.StatusServiceUnavailable, rec.Code)
	}
}
