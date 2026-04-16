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

func TestRateLimiter_Integration(t *testing.T) {
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

	ctx := context.Background()

	// Clean up before test
	_ = client.Del(ctx, "rl:127.0.0.1")

	// Limit of 3 requests per 10 seconds
	limiter := redis.NewRateLimiter(client, 3, 10*time.Second)

	// First 3 should pass
	for i := 0; i < 3; i++ {
		err := limiter.Allow(ctx, "127.0.0.1")
		if err != nil {
			t.Fatalf("request %d should have been allowed: %v", i+1, err)
		}
	}

	// 4th should be rate limited
	err = limiter.Allow(ctx, "127.0.0.1")
	if err == nil {
		t.Fatalf("request 4 should have been rate limited")
	}
}
