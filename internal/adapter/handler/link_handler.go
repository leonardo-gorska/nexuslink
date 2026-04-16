package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leonardo-gorska/nexuslink/internal/domain"
	"github.com/leonardo-gorska/nexuslink/internal/port/input"
)

type LinkHandler struct {
	linkUseCase input.LinkUseCase
}

func NewLinkHandler(linkUseCase input.LinkUseCase) *LinkHandler {
	return &LinkHandler{
		linkUseCase: linkUseCase,
	}
}

type CreateLinkRequest struct {
	URL       string `json:"url"`
	ExpiresIn string `json:"expires_in,omitempty"`
}

type CreateLinkResponse struct {
	Hash        string     `json:"hash"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "https://nexuslink.dev/errors/invalid-request", "Invalid Request", "Malformed JSON body", r.URL.Path, reqID)
		return
	}

	var ttl *time.Duration
	if req.ExpiresIn != "" {
		parsedDuration, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "https://nexuslink.dev/errors/invalid-request", "Invalid Parameter", "Invalid expires_in format", r.URL.Path, reqID)
			return
		}
		ttl = &parsedDuration
	}

	link, err := h.linkUseCase.CreateShortLink(r.Context(), req.URL, ttl)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			RespondError(w, http.StatusBadRequest, "https://nexuslink.dev/errors/invalid-request", "Invalid URL", err.Error(), r.URL.Path, reqID)
			return
		}
		RespondError(w, http.StatusInternalServerError, "https://nexuslink.dev/errors/internal-error", "Internal Error", "Could not create short link", r.URL.Path, reqID)
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	shortURL := scheme + "://" + host + "/r/" + link.Hash

	RespondJSON(w, http.StatusCreated, CreateLinkResponse{
		Hash:        link.Hash,
		ShortURL:    shortURL,
		OriginalURL: link.OriginalURL,
		CreatedAt:   link.CreatedAt,
		ExpiresAt:   link.ExpiresAt,
	})
}

type GetLinkResponse struct {
	Hash        string     `json:"hash"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ClickCount  int64      `json:"click_count"`
	IsActive    bool       `json:"is_active"`
}

func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	hash := chi.URLParam(r, "hash")

	link, err := h.linkUseCase.GetLinkDetails(r.Context(), hash)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			RespondError(w, http.StatusNotFound, "https://nexuslink.dev/errors/link-not-found", "Link Not Found", "No active link found", r.URL.Path, reqID)
			return
		}
		RespondError(w, http.StatusInternalServerError, "https://nexuslink.dev/errors/internal-error", "Internal Error", "Could not retrieve link details", r.URL.Path, reqID)
		return
	}

	RespondJSON(w, http.StatusOK, GetLinkResponse{
		Hash:        link.Hash,
		OriginalURL: link.OriginalURL,
		CreatedAt:   link.CreatedAt,
		ExpiresAt:   link.ExpiresAt,
		ClickCount:  link.ClickCount,
		IsActive:    link.IsActive,
	})
}

func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	hash := chi.URLParam(r, "hash")

	err := h.linkUseCase.DeleteLink(r.Context(), hash)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			RespondError(w, http.StatusNotFound, "https://nexuslink.dev/errors/link-not-found", "Link Not Found", "No active link found", r.URL.Path, reqID)
			return
		}
		RespondError(w, http.StatusInternalServerError, "https://nexuslink.dev/errors/internal-error", "Internal Error", "Could not delete link", r.URL.Path, reqID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
