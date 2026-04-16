package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/app"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

type mockAnalyticsStore struct {
	batchInsertFn func(ctx context.Context, batch []entity.ClickEvent) error
	upsertFn      func(ctx context.Context, stats []entity.DailyAnalytics) error
	queryFn       func(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error)
}

func (m *mockAnalyticsStore) BatchInsertEvents(ctx context.Context, batch []entity.ClickEvent) error {
	if m.batchInsertFn != nil {
		return m.batchInsertFn(ctx, batch)
	}
	return nil
}

func (m *mockAnalyticsStore) UpsertDailyStats(ctx context.Context, stats []entity.DailyAnalytics) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, stats)
	}
	return nil
}

func (m *mockAnalyticsStore) Query(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, hash, from, to)
	}
	return nil, nil
}

func TestAnalyticsService_ProcessClickBatch(t *testing.T) {
	mockStore := &mockAnalyticsStore{
		batchInsertFn: func(ctx context.Context, batch []entity.ClickEvent) error {
			if len(batch) != 1 {
				t.Fatalf("expected 1 element, got %d", len(batch))
			}
			return nil
		},
		upsertFn: func(ctx context.Context, stats []entity.DailyAnalytics) error {
			if len(stats) != 1 {
				t.Fatalf("expected 1 daily stat, got %d", len(stats))
			}
			return nil
		},
	}
	// No geo reader since testing basic processing
	s := app.NewAnalyticsService(mockStore, nil)
	
	err := s.ProcessClickBatch(context.Background(), []entity.ClickEvent{
		{LinkHash: "aB3xK9z", IP: "127.0.0.1", ClickedAt: time.Now()},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
