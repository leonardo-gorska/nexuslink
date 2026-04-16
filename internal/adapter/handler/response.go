package handler

import (
	"encoding/json"
	"net/http"

	"github.com/leonardo-gorska/nexuslink/pkg/httputil"
)

func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func RespondError(w http.ResponseWriter, status int, errorType, title, detail, instance, reqID string) {
	httputil.WriteProblem(w, httputil.ProblemDetail{
		Type:      errorType,
		Title:     title,
		Status:    status,
		Detail:    detail,
		Instance:  instance,
		RequestID: reqID,
	})
}
