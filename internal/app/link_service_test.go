package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/app"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

type mockLinkRepo struct {
	saveFn       func(ctx context.Context, link *entity.Link) error
	findByHashFn func(ctx context.Context, hash string) (*entity.Link, error)
	softDeleteFn func(ctx context.Context, hash string) error
}

func (m *mockLinkRepo) Save(ctx context.Context, link *entity.Link) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, link)
	}
	return nil
}
func (m *mockLinkRepo) FindByHash(ctx context.Context, hash string) (*entity.Link, error) {
	if m.findByHashFn != nil {
		return m.findByHashFn(ctx, hash)
	}
	return nil, nil
}
func (m *mockLinkRepo) SoftDelete(ctx context.Context, hash string) error { return nil }

type mockCacheStore struct {
	getFn func(ctx context.Context, key string) (string, error)
	setFn func(ctx context.Context, key string, value string, ttl time.Duration) error
}

func (m *mockCacheStore) Get(ctx context.Context, key string) (string, error) {
	if m.getFn != nil {
		return m.getFn(ctx, key)
	}
	return "", nil
}
func (m *mockCacheStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, key, value, ttl)
	}
	return nil
}
func (m *mockCacheStore) Delete(ctx context.Context, key string) error { return nil }

func TestLinkService_CreateShortLink(t *testing.T) {
	mockRepo := &mockLinkRepo{
		saveFn: func(ctx context.Context, link *entity.Link) error {
			return nil // simulate success
		},
	}
	s := app.NewLinkService(mockRepo, nil, 7)

	t.Run("success", func(t *testing.T) {
		link, err := s.CreateShortLink(context.Background(), "https://github.com", nil)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if link == nil || link.Hash == "" {
			t.Fatalf("expected valid link")
		}
	})
}

func TestLinkService_ResolveLink(t *testing.T) {
	mockRepo := &mockLinkRepo{
		findByHashFn: func(ctx context.Context, hash string) (*entity.Link, error) {
			return &entity.Link{OriginalURL: "https://github.com"}, nil
		},
	}
	s := app.NewLinkService(mockRepo, nil, 7)

	t.Run("resolve from repo", func(t *testing.T) {
		url, err := s.ResolveLink(context.Background(), "aB3xK9z")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if url != "https://github.com" {
			t.Fatalf("expected https://github.com, got %s", url)
		}
	})
}
