//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/repository/postgres"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

func TestLinkRepo_Integration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nexus:nexus_secret@localhost:5432/nexuslink?sslmode=disable"
	}

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewLinkRepository(pool)
	
	// Create Link
	link := &entity.Link{
		Hash:        "intTest1",
		OriginalURL: "https://integration.test",
		CreatedAt:   time.Now(),
		ClickCount:  0,
		IsActive:    true,
	}

	err = repo.Save(context.Background(), link)
	if err != nil {
		t.Fatalf("failed to save link: %v", err)
	}

	// Find By Hash
	found, err := repo.FindByHash(context.Background(), "intTest1")
	if err != nil {
		t.Fatalf("failed to find link: %v", err)
	}
	if found.OriginalURL != "https://integration.test" {
		t.Fatalf("expected https://integration.test, got %s", found.OriginalURL)
	}

	// Soft Delete
	err = repo.SoftDelete(context.Background(), "intTest1")
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// Find after soft delete (should fail or return nil depending on implementation, but typically our FindByHash checks `is_active=TRUE`)
	_, err = repo.FindByHash(context.Background(), "intTest1")
	if err == nil {
		t.Fatalf("expected error finding deleted link")
	}
}
