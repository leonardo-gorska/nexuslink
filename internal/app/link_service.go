package app

import (
	"context"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
	"github.com/leonardo-gorska/nexuslink/internal/domain/valueobject"
	"github.com/leonardo-gorska/nexuslink/internal/port/output"
	"github.com/leonardo-gorska/nexuslink/pkg/metrics"
)

type LinkService struct {
	repo       output.LinkRepository
	cache      output.CacheStore
	hashLength int
}

func NewLinkService(repo output.LinkRepository, cache output.CacheStore, hashLength int) *LinkService {
	if hashLength == 0 {
		hashLength = 7
	}
	return &LinkService{
		repo:       repo,
		cache:      cache,
		hashLength: hashLength,
	}
}

func (s *LinkService) CreateShortLink(ctx context.Context, originalURL string, ttl *time.Duration) (*entity.Link, error) {
	parsedURL, err := valueobject.NewURL(originalURL)
	if err != nil {
		return nil, err
	}

	var expiresAt *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresAt = &t
	}

	var newLink *entity.Link
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		hash, err := valueobject.NewHash(s.hashLength)
		if err != nil {
			return nil, err
		}

		newLink = &entity.Link{
			Hash:        hash,
			OriginalURL: parsedURL.Raw,
			CreatedAt:   time.Now(),
			ExpiresAt:   expiresAt,
			ClickCount:  0,
			IsActive:    true,
		}

		err = s.repo.Save(ctx, newLink)
		if err == nil {
			return newLink, nil // Success
		}
		// If it's not a collision, or other error, retry might fail too, but we let repository logic define ErrHashCollision
	}

	return nil, domain.ErrHashCollision
}

func (s *LinkService) ResolveLink(ctx context.Context, hash string) (string, error) {
	cacheKey := "link:" + hash
	if s.cache != nil {
		if cachedURL, err := s.cache.Get(ctx, cacheKey); err == nil {
			metrics.CacheHitsTotal.Inc()
			return cachedURL, nil
		}
	}

	metrics.CacheMissesTotal.Inc()
	link, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return "", err
	}

	if link.IsExpired() {
		return "", domain.ErrLinkExpired
	}

	if s.cache != nil {
		ttl := time.Hour // Default 1 hour TTL
		if link.ExpiresAt != nil {
			timeUntilExpiry := time.Until(*link.ExpiresAt)
			if timeUntilExpiry < ttl {
				ttl = timeUntilExpiry
			}
		}
		if ttl > 0 {
			_ = s.cache.Set(ctx, cacheKey, link.OriginalURL, ttl)
		}
	}

	return link.OriginalURL, nil
}

func (s *LinkService) GetLinkDetails(ctx context.Context, hash string) (*entity.Link, error) {
	link, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, hash string) error {
	if err := s.repo.SoftDelete(ctx, hash); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "link:"+hash)
	}
	return nil
}
