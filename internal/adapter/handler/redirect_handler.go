package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leonardo-gorska/nexuslink/internal/domain"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
	"github.com/leonardo-gorska/nexuslink/internal/port/input"
	"github.com/leonardo-gorska/nexuslink/internal/port/output"
)

type RedirectHandler struct {
	linkUseCase input.LinkUseCase
	publisher   output.EventPublisher
}

func NewRedirectHandler(linkUseCase input.LinkUseCase, publisher output.EventPublisher) *RedirectHandler {
	return &RedirectHandler{
		linkUseCase: linkUseCase,
		publisher:   publisher,
	}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	reqID := r.Header.Get("X-Request-ID")

	originalURL, err := h.linkUseCase.ResolveLink(r.Context(), hash)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			RespondError(w, http.StatusNotFound, "https://nexuslink.dev/errors/link-not-found", "Link Not Found", "No active link found", r.URL.Path, reqID)
			return
		}
		if errors.Is(err, domain.ErrLinkExpired) {
			RespondError(w, http.StatusGone, "https://nexuslink.dev/errors/link-expired", "Link Expired", "This link has expired", r.URL.Path, reqID)
			return
		}
		RespondError(w, http.StatusInternalServerError, "https://nexuslink.dev/errors/internal-error", "Internal Error", "Could not resolve link", r.URL.Path, reqID)
		return
	}

	// Fire and forget event publishing
	if h.publisher != nil {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}

		event := &entity.ClickEvent{
			// event_id generation isn't natively in struct, but typical in RMQ
			LinkHash:  hash,
			IP:        ip,
			UserAgent: r.UserAgent(),
			Referer:   r.Referer(),
			ClickedAt: time.Now(),
		}

		eventCtx := context.Background() // Use background context because the request context will be canceled
		go func(e *entity.ClickEvent) {
			// In a real production system, error handling could log failure metrics
			_ = h.publisher.Publish(eventCtx, e)
		}(event)
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}
