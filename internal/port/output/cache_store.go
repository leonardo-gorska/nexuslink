package output

import (
	"context"
	"time"
)

// CacheStore defines the output port for caching logic.
type CacheStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
