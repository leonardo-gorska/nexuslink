package output

import (
	"context"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

// LinkRepository defines the output port for link persistence.
type LinkRepository interface {
	Save(ctx context.Context, link *entity.Link) error
	FindByHash(ctx context.Context, hash string) (*entity.Link, error)
	SoftDelete(ctx context.Context, hash string) error
}
