package input

import (
	"context"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

// LinkUseCase defines the input port for link-related operations.
type LinkUseCase interface {
	CreateShortLink(ctx context.Context, originalURL string, ttl *time.Duration) (*entity.Link, error)
	ResolveLink(ctx context.Context, hash string) (string, error)
	GetLinkDetails(ctx context.Context, hash string) (*entity.Link, error)
	DeleteLink(ctx context.Context, hash string) error
}
