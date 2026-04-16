package redis

import (
	"context"
	"errors"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

type CacheStore struct {
	client *redisclient.Client
}

func NewCacheStore(client *redisclient.Client) *CacheStore {
	return &CacheStore{
		client: client,
	}
}

func (s *CacheStore) Get(ctx context.Context, key string) (string, error) {
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redisclient.Nil) {
		return "", domain.ErrCacheMiss
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *CacheStore) Set(ctx context.Context, key string, val string, ttl time.Duration) error {
	return s.client.Set(ctx, key, val, ttl).Err()
}

func (s *CacheStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
