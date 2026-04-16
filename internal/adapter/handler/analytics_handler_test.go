package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/handler"
	"github.com/leonardo-gorska/nexuslink/internal/domain"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

type mockAnalyticsUseCase struct {
	processFn func(ctx context.Context, batch []entity.ClickEvent) error
	getFn     func(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error)
}

func (m *mockAnalyticsUseCase) ProcessClickBatch(ctx context.Context, batch []entity.ClickEvent) error {
	if m.processFn != nil {
		return m.processFn(ctx, batch)
	}
	return nil
}

func (m *mockAnalyticsUseCase) GetLinkAnalytics(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error) {
	if m.getFn != nil {
		return m.getFn(ctx, hash, from, to)
	}
	return nil, nil
}

func TestGetLinkAnalytics(t *testing.T) {
	tests := []struct {
		name         string
		hashParams   string
		mockReturn   *entity.AnalyticsReport
		mockErr      error
		expectedCode int
	}{
		{
			name:       "Success",
			hashParams: "aB3xK9z",
			mockReturn: &entity.AnalyticsReport{
				Hash:           "aB3xK9z",
				TotalClicks:    100,
				UniqueVisitors: 80,
			},
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Not found",
			hashParams:   "notfoun",
			mockReturn:   nil,
			mockErr:      domain.ErrLinkNotFound,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAnalyticsUseCase{
				getFn: func(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error) {
					if hash == tt.hashParams {
						return tt.mockReturn, tt.mockErr
					}
					return nil, domain.ErrLinkNotFound
				},
			}

			h := handler.NewAnalyticsHandler(mockUC)
			router := chi.NewRouter()
			router.Get("/api/v1/links/{hash}/stats", h.GetLinkAnalytics)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+tt.hashParams+"/stats", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected status %d; got %d", tt.expectedCode, rec.Code)
			}
		})
	}
}
