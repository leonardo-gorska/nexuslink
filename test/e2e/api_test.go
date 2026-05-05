package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/adapter/handler"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
	"github.com/leonardo-gorska/nexuslink/internal/port/input"
)

type mockLinkService struct {
	links map[string]*entity.Link
}

func (m *mockLinkService) CreateShortLink(ctx context.Context, originalURL string, ttl *time.Duration) (*entity.Link, error) {
	if originalURL == "" {
		return nil, domain.ErrInvalidURL
	}
	hash := "aB3xK9z"
	link := &entity.Link{
		Hash:        hash,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
	m.links[hash] = link
	return link, nil
}

func (m *mockLinkService) ResolveLink(ctx context.Context, hash string) (string, error) {
	link, ok := m.links[hash]
	if !ok || !link.IsActive {
		return "", domain.ErrLinkNotFound
	}
	return link.OriginalURL, nil
}

func (m *mockLinkService) GetLinkDetails(ctx context.Context, hash string) (*entity.Link, error) {
	link, ok := m.links[hash]
	if !ok || !link.IsActive {
		return nil, domain.ErrLinkNotFound
	}
	return link, nil
}

func (m *mockLinkService) DeleteLink(ctx context.Context, hash string) error {
	link, ok := m.links[hash]
	if !ok {
		return domain.ErrLinkNotFound
	}
	link.IsActive = false
	return nil
}

type mockRateLimiter struct{}

func (m *mockRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}

func setupTestRouter() *httptest.Server {
	mockService := &mockLinkService{links: make(map[string]*entity.Link)}
	router := handler.NewRouter(
		handler.NewLinkHandler(mockService),
		handler.NewRedirectHandler(mockService, nil),
		handler.NewHealthHandler(),
		handler.NewAnalyticsHandler(nil),
		&mockRateLimiter{},
	)
	return httptest.NewServer(router)
}

func TestEndToEndAPI(t *testing.T) {
	ts := setupTestRouter()
	defer ts.Close()

	// 1. Create a link
	client := ts.Client()
	reqBody := []byte(`{"url":"https://github.com/leonardo-gorska/nexuslink"}`)
	res, err := client.Post(ts.URL+"/api/v1/links", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201, got %d", res.StatusCode)
	}
	var created handler.CreateLinkResponse
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	res.Body.Close()

	if created.Hash != "aB3xK9z" {
		t.Fatalf("Expected hash aB3xK9z, got %s", created.Hash)
	}

	// 2. Get Link Details
	res, err = client.Get(ts.URL + "/api/v1/links/" + created.Hash)
	if err != nil {
		t.Fatalf("Failed to get link: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", res.StatusCode)
	}
	res.Body.Close()

	// 3. Redirect
	// Disable client auto redirect mapping
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	res, err = client.Get(ts.URL + "/r/" + created.Hash)
	if err != nil {
		t.Fatalf("Failed to redirect: %v", err)
	}
	if res.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("Expected 301, got %d", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); loc != "https://github.com/leonardo-gorska/nexuslink" {
		t.Fatalf("Expected location github, got %s", loc)
	}
	res.Body.Close()

	// 4. Delete the link
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/links/"+created.Hash, nil)
	res, err = ts.Client().Do(req)
	if err != nil {
		t.Fatalf("Failed to delete link: %v", err)
	}
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	// 5. Verify it's gone
	res, err = client.Get(ts.URL + "/api/v1/links/" + created.Hash)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", res.StatusCode)
	}
	res.Body.Close()
}
