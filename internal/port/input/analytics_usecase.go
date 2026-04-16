package input

import (
	"context"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

// AnalyticsUseCase defines the input port for analytics operations.
type AnalyticsUseCase interface {
	ProcessClickBatch(ctx context.Context, batch []entity.ClickEvent) error
	GetLinkAnalytics(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error)
}
