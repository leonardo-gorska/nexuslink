package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) BatchInsertEvents(ctx context.Context, events []entity.ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	rows := make([][]interface{}, 0, len(events))
	for _, ev := range events {
		rows = append(rows, []interface{}{
			ev.LinkHash,
			ev.IP,
			ev.UserAgent,
			ev.Referer,
			ev.Country,
			ev.Device,
			ev.Browser,
			ev.OS,
			ev.ClickedAt,
		})
	}

	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"click_events"},
		[]string{"link_hash", "ip_address", "user_agent", "referer", "country", "device_type", "browser", "os", "clicked_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy from failed: %w", err)
	}
	return nil
}

func (r *AnalyticsRepository) UpsertDailyStats(ctx context.Context, stats []entity.DailyAnalytics) error {
	if len(stats) == 0 {
		return nil
	}

	b := &pgx.Batch{}

	query := `
		INSERT INTO daily_analytics (link_hash, event_date, country, device_type, browser, click_count, unique_ips, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (link_hash, event_date, COALESCE(country, ''), COALESCE(device_type, ''), COALESCE(browser, ''))
		DO UPDATE SET 
			click_count = daily_analytics.click_count + EXCLUDED.click_count,
			unique_ips = daily_analytics.unique_ips + EXCLUDED.unique_ips,
			updated_at = NOW()
	`

	for _, s := range stats {
		var c, d, br *string
		if s.Country != "" { c = &s.Country }
		if s.Device != "" { d = &s.Device }
		if s.Browser != "" { br = &s.Browser }

		b.Queue(query, s.LinkHash, s.Date, c, d, br, s.ClickCount, s.UniqueIPs)
	}

	br := r.pool.SendBatch(ctx, b)
	defer br.Close()

	if _, err := br.Exec(); err != nil {
		return fmt.Errorf("batch upsert failed: %w", err)
	}

	return nil
}

func (r *AnalyticsRepository) Query(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error) {
    // Basic stub for analytics query logic to complete interface definition
    return &entity.AnalyticsReport{}, nil
}
