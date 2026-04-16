package output

import (
	"context"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

// AnalyticsStore defines the output port for analytics persistence.
type AnalyticsStore interface {
	BatchInsertEvents(ctx context.Context, events []entity.ClickEvent) error
	UpsertDailyStats(ctx context.Context, stats []entity.DailyAnalytics) error
	Query(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error)
}
