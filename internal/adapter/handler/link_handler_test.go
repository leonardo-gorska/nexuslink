package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leonardo-gorska/nexuslink/internal/adapter/handler"
	"github.com/leonardo-gorska/nexuslink/internal/domain"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

type mockLinkUseCase struct {
	createFn  func(ctx context.Context, originalURL string, ttl *time.Duration) (*entity.Link, error)
	resolveFn func(ctx context.Context, hash string) (string, error)
	getFn     func(ctx context.Context, hash string) (*entity.Link, error)
	deleteFn  func(ctx context.Context, hash string) error
}

func (m *mockLinkUseCase) CreateShortLink(ctx context.Context, originalURL string, ttl *time.Duration) (*entity.Link, error) {
	if m.createFn != nil {
		return m.createFn(ctx, originalURL, ttl)
	}
	return nil, errors.New("not implemented")
}

func (m *mockLinkUseCase) ResolveLink(ctx context.Context, hash string) (string, error) {
	if m.resolveFn != nil {
		return m.resolveFn(ctx, hash)
	}
	return "", errors.New("not implemented")
}

func (m *mockLinkUseCase) GetLinkDetails(ctx context.Context, hash string) (*entity.Link, error) {
	if m.getFn != nil {
		return m.getFn(ctx, hash)
	}
	return nil, errors.New("not implemented")
}

func (m *mockLinkUseCase) DeleteLink(ctx context.Context, hash string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, hash)
	}
	return errors.New("not implemented")
}

func TestCreateLink(t *testing.T) {
	tests := []struct {
		name         string
		reqBody      interface{}
		mockReturn   *entity.Link
		mockErr      error
		expectedCode int
	}{
		{
			name:         "Success",
			reqBody:      map[string]string{"url": "https://github.com"},
			mockReturn:   &entity.Link{Hash: "aB3xK9z", OriginalURL: "https://github.com"},
			mockErr:      nil,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Invalid URL format",
			reqBody:      map[string]string{"url": "invalid-url"},
			mockReturn:   nil,
			mockErr:      domain.ErrInvalidURL,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty URL",
			reqBody:      map[string]string{"url": ""},
			mockReturn:   nil,
			mockErr:      domain.ErrInvalidURL,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockLinkUseCase{
				createFn: func(ctx context.Context, u string, ttl *time.Duration) (*entity.Link, error) {
					return tt.mockReturn, tt.mockErr
				},
			}

			h := handler.NewLinkHandler(mockUC, "http://localhost:8080")
			router := chi.NewRouter()
			router.Post("/api/v1/links", h.CreateLink)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected status %d; got %d", tt.expectedCode, rec.Code)
			}
		})
	}
}

func TestGetLink(t *testing.T) {
	tests := []struct {
		name         string
		hashParams   string
		mockReturn   *entity.Link
		mockErr      error
		expectedCode int
	}{
		{
			name:         "Success",
			hashParams:   "aB3xK9z",
			mockReturn:   &entity.Link{Hash: "aB3xK9z", OriginalURL: "https://github.com"},
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
			mockUC := &mockLinkUseCase{
				getFn: func(ctx context.Context, h string) (*entity.Link, error) {
					if h == tt.hashParams {
						return tt.mockReturn, tt.mockErr
					}
					return nil, errors.New("unexpected hash")
				},
			}

			h := handler.NewLinkHandler(mockUC, "http://localhost:8080")
			router := chi.NewRouter()
			router.Get("/api/v1/links/{hash}", h.GetLink)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+tt.hashParams, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected status %d; got %d", tt.expectedCode, rec.Code)
			}
		})
	}
}

func TestDeleteLink(t *testing.T) {
	tests := []struct {
		name         string
		hashParams   string
		mockErr      error
		expectedCode int
	}{
		{
			name:         "Success",
			hashParams:   "aB3xK9z",
			mockErr:      nil,
			expectedCode: http.StatusNoContent,
		},
		{
			name:         "Not found",
			hashParams:   "notfoun",
			mockErr:      domain.ErrLinkNotFound,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockLinkUseCase{
				deleteFn: func(ctx context.Context, h string) error {
					if h == tt.hashParams {
						return tt.mockErr
					}
					return errors.New("unexpected hash")
				},
			}

			h := handler.NewLinkHandler(mockUC, "http://localhost:8080")
			router := chi.NewRouter()
			router.Delete("/api/v1/links/{hash}", h.DeleteLink)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/"+tt.hashParams, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected status %d; got %d", tt.expectedCode, rec.Code)
			}
		})
	}
}
