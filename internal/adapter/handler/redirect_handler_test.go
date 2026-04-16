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

type mockEventPublisher struct {
	publishFn func(ctx context.Context, e *entity.ClickEvent) error
}

func (m *mockEventPublisher) Publish(ctx context.Context, e *entity.ClickEvent) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, e)
	}
	return nil
}

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name         string
		hashParams   string
		mockReturn   string
		mockErr      error
		expectedCode int
		expectedLoc  string
	}{
		{
			name:         "Success",
			hashParams:   "aB3xK9z",
			mockReturn:   "https://github.com",
			mockErr:      nil,
			expectedCode: http.StatusMovedPermanently,
			expectedLoc:  "https://github.com",
		},
		{
			name:         "Not found",
			hashParams:   "notfoun",
			mockReturn:   "",
			mockErr:      domain.ErrLinkNotFound,
			expectedCode: http.StatusNotFound,
			expectedLoc:  "",
		},
		{
			name:         "Expired",
			hashParams:   "expired",
			mockReturn:   "",
			mockErr:      domain.ErrLinkExpired,
			expectedCode: http.StatusGone,
			expectedLoc:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockLinkUseCase{
				resolveFn: func(ctx context.Context, hash string) (string, error) {
					if hash == tt.hashParams {
						return tt.mockReturn, tt.mockErr
					}
					return "", domain.ErrLinkNotFound
				},
			}

			mockPub := &mockEventPublisher{
				publishFn: func(ctx context.Context, e *entity.ClickEvent) error {
					return nil
				},
			}

			h := handler.NewRedirectHandler(mockUC, mockPub)
			router := chi.NewRouter()
			router.Get("/r/{hash}", h.Redirect)

			req := httptest.NewRequest(http.MethodGet, "/r/"+tt.hashParams, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected status %d; got %d", tt.expectedCode, rec.Code)
			}

			if tt.expectedCode == http.StatusMovedPermanently {
				loc := rec.Header().Get("Location")
				if loc != tt.expectedLoc {
					t.Errorf("expected location %q; got %q", tt.expectedLoc, loc)
				}
			}
		})
	}
}
