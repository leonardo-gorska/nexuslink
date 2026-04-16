package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leonardo-gorska/nexuslink/internal/port/input"
)

type AnalyticsHandler struct {
	analyticsUseCase input.AnalyticsUseCase
}

func NewAnalyticsHandler(analyticsUseCase input.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsUseCase: analyticsUseCase,
	}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// Fase 5 will inject real logic here
	hash := chi.URLParam(r, "hash")
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"hash":            hash,
		"total_clicks":    0,
		"unique_visitors": 0,
		"status":          "under_construction",
	})
}
