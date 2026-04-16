//go:build integration
// +build integration

package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/adapter/repository/redis"
	redisclient "github.com/redis/go-redis/v9"
)

func TestCacheStore_Integration(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opts, err := redisclient.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("failed to parse redis url: %v", err)
	}

	client := redisclient.NewClient(opts)
	defer client.Close()

	store := redis.NewCacheStore(client)
	ctx := context.Background()

	// Clean up before test
	_ = store.Delete(ctx, "intTestCache")

	// Set value
	err = store.Set(ctx, "intTestCache", "cachedValue", 2*time.Second)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Get value
	val, err := store.Get(ctx, "intTestCache")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}
	if val != "cachedValue" {
		t.Fatalf("expected cachedValue, got %s", val)
	}

	// Wait for expiry
	time.Sleep(2500 * time.Millisecond)

	// Get value again, should fail
	_, err = store.Get(ctx, "intTestCache")
	if err == nil {
		t.Fatalf("expected error getting expired cache")
	}
	
	// Delete
	_ = store.Set(ctx, "intTestCache2", "val", 1*time.Minute)
	err = store.Delete(ctx, "intTestCache2")
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
}
