package httputil

import (
	"encoding/json"
	"net/http"
)

// ProblemDetail represents an RFC 7807 Error Response
type ProblemDetail struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteProblem responds with an RFC 7807 Problem Details JSON payload.
func WriteProblem(w http.ResponseWriter, p ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	
	if p.Status == 0 {
		p.Status = http.StatusInternalServerError
	}
	w.WriteHeader(p.Status)

	_ = json.NewEncoder(w).Encode(p)
}
