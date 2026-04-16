package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/leonardo-gorska/nexuslink/pkg/httputil"
)

// Recoverer catches panics, logs the stack trace, and returns a 500 error.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				reqID := GetReqID(r.Context())
				slog.Error("panic recovered",
					slog.String("request_id", reqID),
					slog.Any("error", rvr),
					slog.String("stack", string(debug.Stack())),
				)

				httputil.WriteProblem(w, httputil.ProblemDetail{
					Type:      "https://nexuslink.dev/errors/internal-error",
					Title:     "Internal Error",
					Status:    http.StatusInternalServerError,
					Detail:    "An unexpected error occurred",
					Instance:  r.URL.Path,
					RequestID: reqID,
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
